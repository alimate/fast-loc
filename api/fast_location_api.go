package api

import (
	"context"
	"fmt"
	"time"

	pb "github.com/alimate/fast-loc/gen"
	"github.com/go-redis/redis/v8"
)

const bufferSize = 1000
const bufferTimeout = 15 * time.Millisecond

type FastLocationApi struct {
	redis  *redis.Client
	buffer chan *bufferedRequest
}

func NewFastLocationApi(rdb *redis.Client) *FastLocationApi {
	fls := &FastLocationApi{
		redis:  rdb,
		buffer: make(chan *bufferedRequest, bufferSize),
	}
	go fls.drainBuffer()

	return fls
}

type bufferedRequest struct {
	ctx              context.Context
	location         *pb.DriverLocation
	responseObserver chan error
}

func (api *FastLocationApi) Set(ctx context.Context, location *pb.DriverLocation) (*pb.Empty, error) {
	responseObserver := make(chan error)
	br := &bufferedRequest{
		ctx:              ctx,
		location:         location,
		responseObserver: responseObserver,
	}

	// could block
	api.buffer <- br

	// should block
	err := <-responseObserver
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (api *FastLocationApi) drainBuffer() {
	buffered := make([]*bufferedRequest, 0, bufferSize)
	ticker := time.Tick(bufferTimeout)

	for {
		select {
		case br := <-api.buffer:
			buffered = append(buffered, br)
			if len(buffered) == bufferSize {
				fmt.Println("buffer is full, draining...")
				api.processBuffer(buffered)
				buffered = make([]*bufferedRequest, 0, bufferSize)
			}
		case <-ticker:
			if len(buffered) != 0 {
				fmt.Println("buffer timed out, draining", len(buffered), "items...")
				api.processBuffer(buffered)
				buffered = make([]*bufferedRequest, 0, bufferSize)
			}
		}
	}
}

func (api *FastLocationApi) processBuffer(buffer []*bufferedRequest) {
	if len(buffer) != 0 {
		locations := make([]*redis.GeoLocation, 0, len(buffer))
		for _, br := range buffer {
			locations = append(locations, &redis.GeoLocation{
				Name:      br.location.DriverId,
				Longitude: br.location.Long,
				Latitude:  br.location.Lat,
			})
		}

		err := api.redis.GeoAdd(context.Background(), "fast_driver_locations", locations...).Err()
		for _, br := range buffer {
			br.responseObserver <- err
		}
	}
}

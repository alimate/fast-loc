package api

import (
	"context"

	pb "github.com/alimate/fast-loc/gen"
	"github.com/go-redis/redis/v8"
)

type LocationApi struct {
	redis *redis.Client
}

func (api *LocationApi) Set(ctx context.Context, location *pb.DriverLocation) (*pb.Empty, error) {
	err := api.redis.GeoAdd(ctx, "driver_locations", &redis.GeoLocation{
		Name:      location.DriverId,
		Longitude: location.Long,
		Latitude:  location.Lat,
	}).Err()

	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func NewLocationApi(rdb *redis.Client) *LocationApi {
	return &LocationApi{redis: rdb}
}

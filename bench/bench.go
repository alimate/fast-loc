package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/alimate/fast-loc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func main() {
	var requests int
	flag.IntVar(&requests, "n", 100, "number of requests")
	flag.Parse()

	socket, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatal("failed to connect to gRPC server")
	}
	defer socket.Close()

	client := pb.NewLocationServiceClient(socket)
	wg := &sync.WaitGroup{}
	wg.Add(requests)

	startedAt := time.Now()
	for i := 0; i < requests; i++ {
		name, _ := uuid.NewUUID()
		r := &pb.DriverLocation{
			DriverId: name.String(),
			Lat:      1.5,
			Long:     1.5,
		}

		go func() {
			_, err := client.Set(context.Background(), r)
			if err != nil {
				fmt.Println("failed", err)
			}

			wg.Done()
		}()
	}

	wg.Wait()
	elapsed := time.Since(startedAt)
	fmt.Println("RPS:", float64(requests)/elapsed.Seconds())
	fmt.Println("Millis:", elapsed.Milliseconds())
}

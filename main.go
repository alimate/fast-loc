//go:generate protoc -I proto --go_out=plugins=grpc:gen location_service.proto
package main

import (
	"flag"
	"log"
	"net"

	"github.com/alimate/fast-loc/api"
	pb "github.com/alimate/fast-loc/gen"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

func main() {
	var isFast bool
	flag.BoolVar(&isFast, "fast", false, "should use the fast location service")
	flag.Parse()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:55000",
	})

	socket, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		log.Fatal("failed to listen to port 9000", err)
	}

	server := grpc.NewServer()
	if isFast {
		pb.RegisterLocationServiceServer(server, api.NewFastLocationApi(rdb))
	} else {
		pb.RegisterLocationServiceServer(server, api.NewLocationApi(rdb))
	}

	if err = server.Serve(socket); err != nil {
		log.Fatal("failed serve requests", err)
	}
}

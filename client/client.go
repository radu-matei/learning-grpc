package main

import (
	"flag"
	"log"

	"github.com/radu-matei/learning-grpc/rpc"
	"google.golang.org/grpc"
)

var (
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func main() {
	flag.Parse()

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()

	client := rpc.NewRouteGuideClient(conn)

	printFeature(client, &rpc.Point{Latitude: 409146138, Longitude: -746188906})
	printFeature(client, &rpc.Point{Latitude: 0, Longitude: 0})

	printFeatures(client, &rpc.Rectangle{
		Lo: &rpc.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &rpc.Point{Latitude: 420000000, Longitude: -730000000},
	})

	runRecordRoute(client)

	runRouteChat(client)
}

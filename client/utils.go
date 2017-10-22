package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	"time"

	"github.com/radu-matei/learning-grpc/rpc"
)

// printFeature gets the feature for the given point.
func printFeature(client rpc.RouteGuideClient, point *rpc.Point) {
	log.Printf("Getting feature for point (%d, %d)", point.Latitude, point.Longitude)
	feature, err := client.GetFeature(context.Background(), point)
	if err != nil {
		log.Fatalf("%v.GetFeatures(_) = _, %v: ", client, err)
	}
	log.Println(feature)
}

// printFeatures lists all the features within the given bounding Rectangle.
func printFeatures(client rpc.RouteGuideClient, rect *rpc.Rectangle) {
	log.Printf("Looking for features within %v", rect)
	stream, err := client.ListFeatures(context.Background(), rect)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
	}
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		log.Println(feature)
	}
}

// runRecordRoute sends a sequence of points to server and expects to get a RouteSummary from server.
func runRecordRoute(client rpc.RouteGuideClient) {
	// Create a random number of random points
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	pointCount := int(r.Int31n(100)) + 2 // Traverse at least two points
	var points []*rpc.Point
	for i := 0; i < pointCount; i++ {
		points = append(points, randomPoint(r))
	}
	log.Printf("Traversing %d points.", len(points))
	stream, err := client.RecordRoute(context.Background())
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}
	for _, point := range points {
		if err := stream.Send(point); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, point, err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Printf("Route summary: %v", reply)
}

// runRouteChat receives a sequence of route notes, while sending notes for various locations.
func runRouteChat(client rpc.RouteGuideClient) {
	notes := []*rpc.RouteNote{
		{&rpc.Point{Latitude: 0, Longitude: 1}, "First message"},
		{&rpc.Point{Latitude: 0, Longitude: 2}, "Second message"},
		{&rpc.Point{Latitude: 0, Longitude: 3}, "Third message"},
		{&rpc.Point{Latitude: 0, Longitude: 1}, "Fourth message"},
		{&rpc.Point{Latitude: 0, Longitude: 2}, "Fifth message"},
		{&rpc.Point{Latitude: 0, Longitude: 3}, "Sixth message"},
	}
	stream, err := client.RouteChat(context.Background())
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got message %s at point(%d, %d)", in.Message, in.Location.Latitude, in.Location.Longitude)
		}
	}()
	for _, note := range notes {
		if err := stream.Send(note); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}
	stream.CloseSend()
	<-waitc
}

func randomPoint(r *rand.Rand) *rpc.Point {
	lat := (r.Int31n(180) - 90) * 1e7
	long := (r.Int31n(360) - 180) * 1e7
	return &rpc.Point{Latitude: lat, Longitude: long}
}

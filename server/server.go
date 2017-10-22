package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/radu-matei/learning-grpc/rpc"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "testdata/route_guide_db.json", "A json file containing a list of features")
	port       = flag.Int("port", 10000, "The server port")
)

type routeGuideServer struct {
	savedFeatures []*rpc.Feature
	routeNotes    map[string][]*rpc.RouteNote
}

func (s *routeGuideServer) GetFeature(ctx context.Context, point *rpc.Point) (*rpc.Feature, error) {
	for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}
	f := &rpc.Feature{
		Name:     "",
		Location: point,
	}

	return f, nil
}

func (s *routeGuideServer) ListFeatures(rectangle *rpc.Rectangle, stream rpc.RouteGuide_ListFeaturesServer) error {
	for _, feature := range s.savedFeatures {
		if inRange(feature.Location, rectangle) {
			err := stream.Send(feature)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *routeGuideServer) RecordRoute(stream rpc.RouteGuide_RecordRouteServer) error {

	var pointCount, featureCount, distance int32
	var lastPoint *rpc.Point
	startTime := time.Now()

	for {
		point, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&rpc.RouteSummary{
				PointCount:   pointCount,
				FeatureCount: featureCount,
				Distance:     distance,
				ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}

		pointCount++
		for _, feature := range s.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}

		if lastPoint != nil {
			distance += calcDistance(lastPoint, point)
		}
		lastPoint = point
	}
}

func (s *routeGuideServer) RouteChat(stream rpc.RouteGuide_RouteChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		key := serialize(in.Location)
		for _, note := range s.routeNotes[key] {
			if err := stream.Send(note); err != nil {
				return err
			}
		}
	}
}

func newServer() *routeGuideServer {
	server := new(routeGuideServer)
	server.loadFeatures(*jsonDBFile)
	server.routeNotes = make(map[string][]*rpc.RouteNote)

	return server
}

func main() {
	flag.Parse()

	s := newServer()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("failed to start listening %v", err)
	}
	grpcServer := grpc.NewServer()
	rpc.RegisterRouteGuideServer(grpcServer, s)

	grpcServer.Serve(l)
}

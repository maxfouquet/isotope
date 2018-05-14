package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/Tahler/service-grapher/pkg/graph"
	"istio.io/fortio/fgrpc"
	"istio.io/fortio/fhttp"
)

const port = "8080"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	service, err := getService()
	if err != nil {
		log.Fatal(err)
	}
	err = startServer(service.Type)
	if err != nil {
		log.Fatal(err)
	}
	blockForever()
}

// TODO: This switch should instead be two different programs / images.
func startServer(serviceType graph.ServiceType) (err error) {
	switch serviceType {
	case graph.HTTPService:
		err = startHTTPServer()
	case graph.GRPCService:
		err = startGRPCServer()
	default:
		err = fmt.Errorf("unknown value of service type: %v", serviceType)
	}
	return
}

func startHTTPServer() (err error) {
	_, tcpAddr := fhttp.Serve(port, "")
	if tcpAddr == nil {
		err = errors.New("failed to start HTTP server")
	}
	return
}

func startGRPCServer() (err error) {
	boundPort := fgrpc.PingServer(port, fgrpc.DefaultHealthServiceName, 0)
	if boundPort == -1 {
		err = errors.New("failed to start GRPC server")
	}
	return
}

func blockForever() {
	select {}
}

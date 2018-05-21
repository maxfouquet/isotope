package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/Tahler/service-grapher/pkg/consts"
	"github.com/Tahler/service-grapher/pkg/graph/svc"
	"github.com/Tahler/service-grapher/pkg/graph/svctype"
	"istio.io/fortio/fgrpc"
	"istio.io/fortio/fhttp"
)

func main() {
	log.Printf("Listening on port %v\n", consts.ServicePort)
	startServer(service)
	blockForever()
}

// TODO: This switch should instead be two different programs / images.
func startServer(service svc.Service) (err error) {
	switch service.Type {
	case svctype.ServiceHTTP:
		err = startHTTPServer(service)
	case svctype.ServiceGRPC:
		err = startGRPCServer(service)
	default:
		err = fmt.Errorf("unknown value of service type: %v", service.Type)
	}
	return
}

func startHTTPServer(service svc.Service) (err error) {
	// mux, addr := fhttp.HTTPServer("echo", string(consts.ServicePort))
	// if addr == nil {
	// 	err = fmt.Errorf("")
	// 	return
	// }

	_, tcpAddr := fhttp.Serve(string(consts.ServicePort), "")
	if tcpAddr == nil {
		err = errors.New("failed to start HTTP server")
	}
	return
}

func startGRPCServer(service svc.Service) (err error) {
	boundPort := fgrpc.PingServer(
		string(consts.ServicePort), fgrpc.DefaultHealthServiceName, 0)
	if boundPort == -1 {
		err = errors.New("failed to start GRPC server")
	}
	return
}

func blockForever() {
	select {}
}

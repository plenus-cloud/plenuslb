package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"k8s.io/klog"
	"plenus.io/plenuslb/pkg/operator/server"
	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
	"plenus.io/plenuslb/pkg/utils/k8shealth"
)

func main() {
	klog.InitFlags(nil)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 10000))
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	networkServer := &server.PlenusLbServer{}
	plenuslbV1Alpha1.RegisterPlenusLbServer(grpcServer, networkServer)

	// TODO: must be improved and expore ready probe when is really ready
	go k8shealth.HealthHandlers()

	createStopHandler(grpcServer)

	klog.Infof("Starting grpc server on address %s", lis.Addr().String())

	go func() {
		time.Sleep(2 * time.Second)
		k8shealth.IsReady = true
	}()

	err = grpcServer.Serve(lis)
	if err != nil {
		klog.Fatalf("Error starting gRPC server %s", err.Error())
	}
	klog.Info("gRPC server stopped")
}

func createStopHandler(grpcServer *grpc.Server) {
	// listen for interrupts or the Linux SIGTERM signal and
	// stop gracefully stop grpc server
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		grpcServer.GracefulStop()
	}()
}

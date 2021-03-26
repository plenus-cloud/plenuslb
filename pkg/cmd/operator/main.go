/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	"plenus.io/plenuslb/pkg/operator/observer"
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

	// subscribe to addresses update and watch them 
	observer.Run()

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

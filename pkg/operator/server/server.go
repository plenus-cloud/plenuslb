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

package server

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"

	"plenus.io/plenuslb/pkg/operator/observer"
	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
)

// PlenusLbServer is the grpc server implementation
type PlenusLbServer struct {
}

var myNodeName string

func init() {
	n := os.Getenv("MY_NODE_NAME")
	if n == "" {
		klog.Fatal("MY_NODE_NAME env is required")
	}

	myNodeName = n
}

// HealthProbe is dumb ping-pong
func (s *PlenusLbServer) HealthProbe(ctx context.Context, ping *plenuslbV1Alpha1.Ping) (*plenuslbV1Alpha1.Pong, error) {
	klog.Infof("Received ping request: %s", ping.GetMessage())

	return &plenuslbV1Alpha1.Pong{
		Message: "PONG",
	}, nil
}

// AddAddress adds given address to specific interface
func (s *PlenusLbServer) AddAddress(ctx context.Context, info *plenuslbV1Alpha1.AddressInfo) (*plenuslbV1Alpha1.Result, error) {
	klog.Infof("Received request to add address %s to interface %s", info.GetAddress(), info.GetInterface())

	err := observer.AddAddress(info)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	m := fmt.Sprintf("Added address %s on interface %s of node %s", info.GetAddress(), info.GetInterface(), myNodeName)
	return &plenuslbV1Alpha1.Result{
		Done:    true,
		Message: m,
	}, nil
}

// RemoveAddress removes given address from specific interface
func (s *PlenusLbServer) RemoveAddress(ctx context.Context, info *plenuslbV1Alpha1.AddressInfo) (*plenuslbV1Alpha1.Result, error) {
	klog.Infof("Received request to remove address %s from interface %s", info.GetAddress(), info.GetInterface())

	err := observer.RemoveAddress(info)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	m := fmt.Sprintf("Deleted address %s from interface %s of node %s", info.GetAddress(), info.GetInterface(), myNodeName)
	return &plenuslbV1Alpha1.Result{
		Done:    true,
		Message: m,
	}, nil
}

// Cleanup removes all addresses managed by plenuslb from all interfaces
func (s *PlenusLbServer) Cleanup(ctx context.Context, cleanupInfo *plenuslbV1Alpha1.CleanupInfo) (*plenuslbV1Alpha1.Result, error) {
	toKeep := cleanupInfo.GetKeepThese()
	klog.Infof("Received request to cleanup all addresses managed by plenuslb, except %d", len(toKeep))

	err := observer.Cleanup(toKeep)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &plenuslbV1Alpha1.Result{
		Done:    true,
		Message: "Cleanup completed",
	}, nil
}

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
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
	"k8s.io/klog"

	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/leader"
	"plenus.io/plenuslb/pkg/utils/k8shealth"
)

func init() {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
}

func main() {
	klog.InitFlags(nil)

	klog.Info("Creating clientset")
	clients.CreateClientsOrDie()
	config := clients.GetInClusterConfig()

	lead := leader.Election{}
	id := os.Getenv("MY_POD_NAME")
	if id == "" {
		klog.Fatal("MY_POD_NAME env variable is required!")
	}
	leaseLockNamespace := os.Getenv("MY_POD_NAMESPACE")
	if leaseLockNamespace == "" {
		klog.Fatal("MY_POD_NAMESPACE env variable is required!")
	}

	lead.Init(config, leaseLockNamespace, id)

	// use a Go context so we can tell the leaderelection code when we
	// want to step down
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	createStopHandler(ctx, cancel, config)

	// we need to know when the leader gorootine is ended, send a sessage over a channel when I've done
	leaderStop := make(chan bool)
	// start leader bussiness
	go lead.DoLeaderElection(ctx, leaderStop)

	klog.Infof("Leader election started")
	klog.Info("Waiting for leader release")

	go k8shealth.HealthHandlers()
	k8shealth.IsReady = true
	// wait for the leader goroutine
	<-leaderStop
	klog.Info("Leader have been stopped")
}

func createStopHandler(ctx context.Context, cancel context.CancelFunc, config *rest.Config) {
	// use a client that will stop allowing new requests once the context ends
	config.Wrap(transport.ContextCanceller(ctx, fmt.Errorf("the leader is shutting down")))
	// listen for interrupts or the Linux SIGTERM signal and cancel
	// our context, which the leader election code will observe and
	// step down
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		cancel()
	}()
}

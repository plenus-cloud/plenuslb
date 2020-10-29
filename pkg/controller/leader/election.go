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

package leader

import (
	"context"
	"fmt"
	"os"
	"time"

	apiextensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"
	plenuslbclient "plenus.io/plenuslb/pkg/client/clientset/versioned"
	"plenus.io/plenuslb/pkg/controller/allocationswatcher"
	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/ephemeralips"
	"plenus.io/plenuslb/pkg/controller/events"
	"plenus.io/plenuslb/pkg/controller/ipallocations"
	"plenus.io/plenuslb/pkg/controller/operator"
	"plenus.io/plenuslb/pkg/controller/persistentips"
	poolscontroller "plenus.io/plenuslb/pkg/controller/poolsController"
	"plenus.io/plenuslb/pkg/controller/servicewatcher"
)

// Election manage all the election stuff
type Election struct {
	lock                  *resourcelock.LeaseLock
	clientset             clientset.Interface
	plenuslbclient        plenuslbclient.Interface
	apiextensionclientset *apiextensionclientset.Clientset
	config                *rest.Config
	leaseLockName         string
	leaseLockNamespace    string
	id                    string
	imLeader              bool
}

// ReleaseOnCancel  LeaseLock option seem to be buggy: when the leader is released the configuration (LeaseDuration, RenewDeadline, RetryPeriod) are lost
// le.releaseLeader is the workaround
func (le *Election) releaseLeader() error {
	klog.Infof("Getting lock")
	electionRecord, err := le.lock.Get()
	if err != nil {
		klog.Errorf("Cannot get lock object: %s", err.Error())
		return err
	}

	now := metav1.Now()
	newElectionRecord := *electionRecord
	newElectionRecord.HolderIdentity = ""
	newElectionRecord.RenewTime = now
	klog.Infof("Updating lock")
	err = le.lock.Update(newElectionRecord)
	if err != nil {
		klog.Errorf("Failed to release lock object: %s", err.Error())
		return err
	}
	klog.Info("Lock released")
	return nil
}

// Init intialize the LeaderElection struct
func (le *Election) Init(
	config *rest.Config,
	leaseLockNamespace,
	id string,
) {
	le.clientset = clients.GetK8sClient()
	le.apiextensionclientset = clients.GetExtensionClient()
	le.plenuslbclient = clients.GetPlenuslbClient()
	le.leaseLockName = "plenus.io-leader-lease"
	le.leaseLockNamespace = leaseLockNamespace
	le.id = id
	le.config = config
}

// DoLeaderElection perform the election stuffs and start the leader-bussines if elected
func (le *Election) DoLeaderElection(ctx context.Context, stopped chan bool) {
	// leader election uses the Kubernetes API by writing to a
	// lock object, which can be a LeaseLock object (preferred),
	// a ConfigMap, or an Endpoints (deprecated) object.
	// Conflicting writes are detected and each client handles those actions
	// independently.

	// we use the Lease lock type since edits to Leases are less common
	// and fewer objects in the cluster watch "all Leases".
	le.lock = &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      le.leaseLockName,
			Namespace: le.leaseLockNamespace,
		},
		Client: le.clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: le.id,
		},
	}

	stopCh := make(chan struct{})

	// start the leader election code loop
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: le.lock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.

		// this seem to be buggy: when the leader is released the configuration (LeaseDuration, RenewDeadline, RetryPeriod) are lost
		// le.releaseLeader is the workaround
		ReleaseOnCancel: false,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// we're notified when we start - this is where you would
				// usually put your code
				klog.Infof("I'm the leader(%s), starting leader business", le.id)
				le.imLeader = true

				klog.Info("Creating Allocation Custom Resource Definition")
				err := ipallocations.CreateOrUpdateCRD()
				if err != nil {
					klog.Fatal(err)
					return
				}

				operator.Init()

				klog.Info("Creating PersistentIPPool Custom Resource Definition")
				err = persistentips.CreateOrUpdateCRD()
				if err != nil {
					klog.Fatal(err)
					return
				}
				klog.Info("Starting persistent ippools watcher")
				persistentips.Init()
				persistentips.WatchIPPools(stopCh)

				klog.Info("Creating EphemeralIPPool Custom Resource Definition")
				err = ephemeralips.CreateOrUpdateCRD()
				if err != nil {
					klog.Fatal(err)
					return
				}
				klog.Info("Starting ephemeral ippools watcher")
				ephemeralips.Init()
				ephemeralips.WatchIPPools(stopCh)

				klog.Info("Starting ipallocator")
				servicewatcher.Init()
				servicewatcher.WatchServices(stopCh)
				allocationswatcher.Init()
				allocationswatcher.WatchAllocations(stopCh)

				// Wait for all involved caches to be synced, before processing items from the queue is started
				klog.Infof("Syncing all stores")
				cacheInformers := []cache.InformerSynced{
					allocationswatcher.AllocationController.HasSynced,
					servicewatcher.ServicesController.HasSynced,
					persistentips.IPPoolsController.HasSynced,
					ephemeralips.IPPoolsController.HasSynced,
				}
				if operator.OperatorNodesController != nil {
					cacheInformers = append(cacheInformers, operator.OperatorNodesController.HasSynced)
				}

				poolscontroller.Init()

				if !cache.WaitForCacheSync(
					stopCh,
					cacheInformers...,
				) {
					runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
					return
				}

				klog.Info("########### Plenus LB is ready ###########")

				events.ListenModifiedPersistentPoolsChan(stopCh)
				events.ListenDeletedPersistentPoolsChan(stopCh)
				events.ListenModifiedEphemeralPoolsChan(stopCh)
				events.ListenDeletedEphemeralPoolsChan(stopCh)
				events.ListenOperatorNodeLostChan(stopCh)
				events.ListenNewOperatorNodeChan(stopCh)

				<-stopCh

				klog.Info("Leader bussines exited")
			},
			OnStoppedLeading: func() {
				// we can do cleanup here, or after the RunOrDie method
				// returns
				klog.Infof("I'm not the leader anymore. Stopping leader things")
				if !le.imLeader {
					return
				}
				le.imLeader = false
				close(stopCh)
				// I just got the lock
				klog.Info("Releasing leader lock")
				if err := le.releaseLeader(); err != nil {
					klog.Fatalf("Failed to release leader lock: %s", err.Error())
				}
				klog.Info("Leader lock released")

				// TODO: for now, wen the leading is stopped restart the process to esure a clean situation
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == le.id {
					// I just got the lock
					return
				}
				klog.Infof("new leader elected: %v", identity)
			},
		},
	})
	stopped <- true
}

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

package ephemeralips

import (
	"reflect"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/events"
	"plenus.io/plenuslb/pkg/controller/operator"
	"plenus.io/plenuslb/pkg/controller/utils"
)

var (
	ippoolsStore cache.Store
	// IPPoolsController manages the ippools resources
	IPPoolsController cache.Controller
)

func warmupIPPoolsCacheOrDie() {
	list, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().EphemeralIPPools().List(metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof("Warming-up ephemeral ippools cache, %d ippools are on the cluster now", len(list.Items))
	for _, ippool := range list.Items {
		_ = ippoolsStore.Add(ippool)
	}
}

func createIPPoolsWatcher() {
	watchlist := cache.NewListWatchFromClient(
		clients.GetPlenuslbClient().LoadbalancingV1alpha1().RESTClient(),
		loadbalancing_v1alpha1.EphemeralIPPoolCRDPlural,
		v1.NamespaceAll,
		fields.Everything(),
	)
	store, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		watchlist,
		&loadbalancing_v1alpha1.EphemeralIPPool{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pool, ok := obj.(*loadbalancing_v1alpha1.EphemeralIPPool)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				klog.Infof("Added new ephemeral ippool %s", pool.GetName())
				addPool(pool)
			},
			DeleteFunc: func(obj interface{}) {
				pool, ok := obj.(*loadbalancing_v1alpha1.EphemeralIPPool)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				klog.Infof("Deleted ephemeral ippool %s", pool.GetName())
				removePool(pool)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newPool, ok := newObj.(*loadbalancing_v1alpha1.EphemeralIPPool)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(newObj))
					return
				}
				klog.Infof("Modified ephemeral ippool %s", newPool.GetName())
				modifyPool(newPool)
			},
		},
	)

	ippoolsStore = store
	IPPoolsController = controller

}

// WatchIPPools starts whatching the ippools resources
func WatchIPPools(stop chan struct{}) {
	go IPPoolsController.Run(stop)
}

func addPool(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	// check if pool had addAddressesToInterface options, if yes ensure is daemonset is presence
	if utils.EphemeralPoolHasHostNetworkOption(pool) && !operator.IsDeployed() {
		_ = operator.DeployOrDie()
	}
}

func modifyPool(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	events.EphemeralPoolModified(pool)
}

func removePool(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	events.EphemeralPoolDeleted(pool)
}

// GetPoolsList returns the list of ephemeral ip pools
func GetPoolsList() []interface{} {
	return ippoolsStore.List()
}

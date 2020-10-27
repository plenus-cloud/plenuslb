package persistentips

import (
	"fmt"
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
	list, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().PersistentIPPools().List(metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof("Warming-up ippools cache, %d ippools are on the cluster now", len(list.Items))
	for _, ippool := range list.Items {
		ippoolsStore.Add(ippool)
	}
}

func warmupIPAvailabilityOrDie() error {
	klog.Info("Processing IPs availability")
	allocations, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		klog.Error("Failed to process ip availability")
		klog.Fatal(err)
		return err
	}

	for _, obj := range ippoolsStore.List() {
		pool, ok := obj.(*loadbalancing_v1alpha1.PersistentIPPool)
		if !ok {
			err := fmt.Errorf("unexpected type %s", reflect.TypeOf(obj))
			klog.Error(err)
			continue
		}
		addPool(pool)
		processIPAvailability(pool, allocations)
	}
	return nil
}

func createIPPoolsWatcher() {
	watchlist := cache.NewListWatchFromClient(
		clients.GetPlenuslbClient().LoadbalancingV1alpha1().RESTClient(),
		loadbalancing_v1alpha1.PersistentIPPoolCRDPlural,
		v1.NamespaceAll,
		fields.Everything(),
	)
	store, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		watchlist,
		&loadbalancing_v1alpha1.PersistentIPPool{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pool, ok := obj.(*loadbalancing_v1alpha1.PersistentIPPool)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				klog.Infof("Added new ippool %s", pool.GetName())
				addPool(pool)
			},
			DeleteFunc: func(obj interface{}) {
				pool, ok := obj.(*loadbalancing_v1alpha1.PersistentIPPool)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				klog.Infof("Deleted new ippool %s", pool.GetName())
				removePool(pool)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newPool, ok := newObj.(*loadbalancing_v1alpha1.PersistentIPPool)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(newObj))
					return
				}
				klog.Infof("Modified new ippool %s", newPool.GetName())
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

func addPool(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	addOrReplaceAvailabilityPool(pool.DeepCopy())
	// check if pool had addAddressesToInterface options, if yes ensure is daemonset is presence
	if utils.PersistentPoolHasHostNetworkOption(pool) && !operator.IsDeployed() {
		operator.DeployOrDie()
	}
}

func modifyPool(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	events.PersistentPoolModified(pool)
}

func removePool(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	events.PersistentPoolDeleted(pool)
}

// GetPoolsList returns the list of persistent ip pools
func GetPoolsList() []interface{} {
	return ippoolsStore.List()
}

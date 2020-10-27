package ephemeralips

import (
	"errors"
	"reflect"

	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// ErrPoolNotFound is returned when is requested a non-existing pool
var ErrPoolNotFound = errors.New("No pool available")

var poolStoreList = func() []interface{} {
	return ippoolsStore.List()
}

func getPoolForService(serviceNamespace string) *loadbalancing_v1alpha1.EphemeralIPPool {
	for _, obj := range poolStoreList() {
		pool, ok := obj.(*loadbalancing_v1alpha1.EphemeralIPPool)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			continue
		}

		if len(pool.Spec.AllowedNamespaces) == 0 || utils.ContainsString(pool.Spec.AllowedNamespaces, serviceNamespace) {
			return pool
		}
	}

	return nil
}

// SearchPoolByName get a pool by name
func SearchPoolByName(name string) *loadbalancing_v1alpha1.EphemeralIPPool {
	for _, obj := range poolStoreList() {
		pool, ok := obj.(*loadbalancing_v1alpha1.EphemeralIPPool)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			return nil
		}
		if pool.GetName() == name {
			return pool
		}
	}
	return nil
}

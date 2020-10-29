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

package persistentips

import (
	"errors"
	"reflect"
	"sync"

	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// ErrNoIPAvailable is returned when is requested an IP but none is available
var ErrNoIPAvailable = errors.New("No ip available")

// ErrPoolNotFound is returned when is requested a non-existing pool
var ErrPoolNotFound = errors.New("No pool available")

var availablesPools []*loadbalancing_v1alpha1.PersistentIPPool

var availabilityLock = &sync.Mutex{}

var poolStoreList = func() []interface{} {
	return ippoolsStore.List()
}

// UpdatePoolAvailability syncronize the pool availability with the modified pool
func UpdatePoolAvailability(pool *loadbalancing_v1alpha1.PersistentIPPool, addresses []string) {
	if len(addresses) == 0 {
		return
	}
	klog.Infof("Updating pool availability %s", pool.GetName())
	for _, existingPool := range availablesPools {
		if existingPool.GetName() == pool.GetName() {
			pool.DeepCopyInto(existingPool)
			existingPool.Spec.Addresses = addresses
			return
		}
	}

	klog.Infof("All addresses of pool %s was used, recreating availability", pool.GetName())
	recreatePoolAvailability(pool, addresses)

	logPools()
}

// recreatePoolAvailability reacreate the availability of a pool the was completely used
func recreatePoolAvailability(originalPool *loadbalancing_v1alpha1.PersistentIPPool, addresses []string) {
	availablePool := originalPool.DeepCopy()
	availablePool.Spec.Addresses = addresses
	availablesPools = append(availablesPools, availablePool)
}

// removeAvailablePool remove a given pool from the availability list
func removeAvailablePool(pool *loadbalancing_v1alpha1.PersistentIPPool) bool {
	poolIndex := -1
	klog.Infof("Removing available pool %s from dict", pool.GetName())
	for index, existingPool := range availablesPools {
		if existingPool.GetName() == pool.GetName() {
			poolIndex = index
			break
		}
	}

	if poolIndex == -1 {
		klog.Infof("Pool %s is not into availability list", pool.GetName())
		return false
	}

	// Remove the element at index i from a.
	availablesPools[poolIndex] = availablesPools[len(availablesPools)-1] // Copy lc.poolsst element to index i.
	availablesPools[len(availablesPools)-1] = nil                        // Erc.poolsse lc.poolsst element (write zero vc.poolslue).
	availablesPools = availablesPools[:len(availablesPools)-1]           // Truncc.poolste slice.

	return true
}

func logPools() {
	for _, pool := range availablesPools {
		klog.Infof("--- Pool %s has %d (%v) available addesses  ", pool.GetName(), len(pool.Spec.Addresses), pool.Spec.Addresses)
	}
}

// UseIP remove the ip from the pool availabiliti, only if namespace sontraint are satisfied
func UseIP(namespace, address string) (*loadbalancing_v1alpha1.PersistentIPPool, error) {
	availabilityLock.Lock()
	defer availabilityLock.Unlock()
	klog.Infof("Cheching if address %s is usable for namespace %s", address, namespace)
	logPools()

	if len(availablesPools) == 0 {
		return nil, ErrNoIPAvailable
	}

	pool := getAvailabilityPoolOfAddress(namespace, address)

	if pool == nil || len(pool.Spec.Addresses) == 0 {
		klog.Warningf("Ip %s is not available for namespace %s", address, namespace)
		return nil, ErrNoIPAvailable
	}
	klog.Infof("Available addresses in pool %s are %v", pool.GetName(), pool.Spec.Addresses)
	// Remove the element at index i from a.
	removeIPFromPool(pool, address)

	klog.Infof("Address %s is usable for namespace %s and belongs to pool %s ", address, namespace, pool.GetName())
	return pool, nil
}

func removeIPFromPool(pool *loadbalancing_v1alpha1.PersistentIPPool, address string) {
	addresses := []string{}
	for _, a := range pool.Spec.Addresses {
		if a != address {
			addresses = append(addresses, a)
		}
	}

	pool.Spec.Addresses = addresses
	if len(pool.Spec.Addresses) == 0 {
		removeAvailablePool(pool)
	}
}

func searchAvailabilityPoolByName(name string) *loadbalancing_v1alpha1.PersistentIPPool {
	for _, existingPool := range availablesPools {
		if existingPool.GetName() == name {
			return existingPool
		}
	}
	return nil
}

// SearchPoolByName get a pool by name
func SearchPoolByName(name string) *loadbalancing_v1alpha1.PersistentIPPool {
	for _, obj := range poolStoreList() {
		pool, ok := obj.(*loadbalancing_v1alpha1.PersistentIPPool)
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

// ProcessIPAvailability calulates and caches the available addresses from allocations list and pools list
func processIPAvailability(pool *loadbalancing_v1alpha1.PersistentIPPool, allocations *loadbalancing_v1alpha1.IPAllocationList) {
	availabilityLock.Lock()
	defer availabilityLock.Unlock()
	klog.Infof("Processing IPs availability of pool %s", pool.GetName())

	freePool := pool.DeepCopy()
	freePool.Spec.Addresses = []string{}
	for _, ip := range pool.Spec.Addresses {
		if !checkIfIPIsAllocated(ip, allocations) {
			klog.Infof("IP %s of pool %s is available", ip, pool.GetName())
			freePool.Spec.Addresses = append(freePool.Spec.Addresses, ip)
		}
	}

	addOrReplaceAvailabilityPool(freePool)
}

// ProcessIPAvailabilityFromCacheList calulates and caches the available addresses from allocations list and pools list
func ProcessIPAvailabilityFromCacheList(pool *loadbalancing_v1alpha1.PersistentIPPool, allocations []interface{}) {
	availabilityLock.Lock()
	defer availabilityLock.Unlock()
	klog.Infof("Processing IPs availability of pool %s", pool.GetName())

	freePool := pool.DeepCopy()
	freePool.Spec.Addresses = []string{}
	for _, ip := range pool.Spec.Addresses {
		if !checkIfIPIsAllocatedByCache(ip, allocations) {
			klog.Infof("IP %s of pool %s is available", ip, pool.GetName())
			freePool.Spec.Addresses = append(freePool.Spec.Addresses, ip)
		}
	}

	addOrReplaceAvailabilityPool(freePool)
}

func addOrReplaceAvailabilityPool(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	defer logPools()
	if len(pool.Spec.Addresses) > 0 {
		existingPool := searchAvailabilityPoolByName(pool.GetName())
		if existingPool != nil {
			*existingPool = *pool
			klog.Infof("Replacing pool %s with new availability", pool.GetName())
		} else {
			availablesPools = append(availablesPools, pool)
			klog.Infof("Adding new pool %s to availability", pool.GetName())
		}
	} else {
		klog.Infof("No availability for pool %s", pool.GetName())
	}
}

func checkIfIPIsAllocatedByCache(ip string, allocations []interface{}) bool {
	for _, obj := range allocations {
		allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			continue
		}
		for _, addrAlloc := range allocation.Spec.Allocations {
			if addrAlloc.Address == ip {
				return true
			}
		}
	}
	return false
}

func checkIfIPIsAllocated(ip string, allocations *loadbalancing_v1alpha1.IPAllocationList) bool {
	for _, allocation := range allocations.Items {
		for _, addrAlloc := range allocation.Spec.Allocations {
			if addrAlloc.Address == ip {
				return true
			}
		}
	}
	return false
}

// GetPoolOfAddress returns the pool of a given address, filtered by namespace name
func GetPoolOfAddress(namespace, address string) *loadbalancing_v1alpha1.PersistentIPPool {
	for _, obj := range poolStoreList() {
		pool, ok := obj.(*loadbalancing_v1alpha1.PersistentIPPool)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			continue
		}

		if len(pool.Spec.AllowedNamespaces) == 0 || utils.ContainsString(pool.Spec.AllowedNamespaces, namespace) {
			for _, poolIP := range pool.Spec.Addresses {
				if poolIP == address {
					return pool
				}
			}
		}
	}

	return nil
}

func getAvailabilityPoolOfAddress(namespace, address string) *loadbalancing_v1alpha1.PersistentIPPool {
	for _, pool := range availablesPools {
		if len(pool.Spec.AllowedNamespaces) == 0 || utils.ContainsString(pool.Spec.AllowedNamespaces, namespace) {
			for _, poolIP := range pool.Spec.Addresses {
				if poolIP == address {
					return pool
				}
			}
		}
	}

	return nil
}

// EnsureAddressIsNotAvailable ensure that a give ip in not in the availability list
func EnsureAddressIsNotAvailable(poolName, namespace, address string) {
	klog.Infof("Ensuring IP %s of pool is not available", address)
	availabilityPool := searchAvailabilityPoolByName(poolName)
	if availabilityPool == nil {
		pool := GetPoolOfAddress(namespace, address)
		if pool != nil && pool.GetName() != poolName {
			klog.Errorf("Address %s is of pool %s, not of pool %s. If you have manually created an allocation, please fix it ", address, pool.GetName(), poolName)
			return
		} else if pool == nil {
			klog.Errorf("Pool of address %s does not exists", address)
			return
		} else {
			klog.Infof("Address %s of pool %s is not available", address, pool.GetName())
			return
		}
	}

	// Remove the element
	removeIPFromPool(availabilityPool, address)
}

// ReleaseIP restore ad ip into the pool availability
func ReleaseIP(poolName, namespace, address string) {
	availabilityLock.Lock()
	defer availabilityLock.Unlock()
	klog.Infof("Releasing IP %s of pool %s", address, poolName)
	availabilityPool := searchAvailabilityPoolByName(poolName)
	// maybe the pool exists but all addresses are used
	if availabilityPool == nil {
		pool := GetPoolOfAddress(namespace, address)
		if pool == nil {
			klog.Warningf("Cannot find pool of address %s", address)
			return
		}
		// all addresses of pool are used, recreate availability
		recreatePoolAvailability(pool, []string{address})
	} else {
		for _, poolAddress := range availabilityPool.Spec.Addresses {
			if poolAddress == address {
				klog.Errorf("IP %s of pool %s is already available", address, poolName)
				return
			}
		}
		pool := GetPoolOfAddress(namespace, address)
		if pool == nil {
			klog.Warningf("Cannot find pool of address %s", address)
			return
		}
		availabilityPool.Spec.Addresses = append(availabilityPool.Spec.Addresses, address)
	}

	logPools()
}

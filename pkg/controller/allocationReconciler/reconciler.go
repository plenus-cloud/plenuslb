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

package allocationreconciler

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/ephemeralips"
	"plenus.io/plenuslb/pkg/controller/ipallocations"
	"plenus.io/plenuslb/pkg/controller/persistentips"
	"plenus.io/plenuslb/pkg/controller/servicesupdater"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// ReconcileAllocation ensures consistency between service and IPAllocation
func ReconcileAllocation(service *v1.Service, allocation *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	klog.Infof("Reconciling allocation %s/%s with the service", allocation.GetNamespace(), allocation.GetName())
	if !utils.ServiceIsLoadBalancer(service) {
		return nil, serviceIsNoLongerALoadBalancer(service, allocation)
	}

	expectedAllocationType := expectedAllocationType(service)
	if allocation.Spec.Type != expectedAllocationType {
		return changeAllocationType(service, allocation)
	}

	if expectedAllocationType == loadbalancing_v1alpha1.PersistentIP {
		return reconcilePersistentAllocation(service, allocation)
	}

	if expectedAllocationType == loadbalancing_v1alpha1.EphemeralIP {
		return reconcileEphemeralAllocation(service, allocation)
	}
	return nil, nil
}

//	-> has externals ips?
//		-> if yes:
//			-> check the current allocation as the same ips allocated
//			-> update the allocation and set as pending (allocator will ensure allocations)
//			-> deallocate and release removed ips
//		-> if no: why? if is persistent it should
func reconcilePersistentAllocation(service *v1.Service, allocation *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	isPersistentAllocation, ips := utils.IsPersistentAllocation(service)

	if isPersistentAllocation {
		patched, removedAllocations, allocErr, err := persistentips.CheckAndPatchAllocation(service.GetNamespace(), service.GetName(), ips, allocation)
		if err != nil {
			return nil, err
		}

		// THIS IS VERY IMPORTANT!
		// I do cleanup operations only if the allocation was is success state.
		// This because otherwise I risk to deallocate address of other service/allocation
		if allocation.Status.State == loadbalancing_v1alpha1.AllocationStatusSuccess {
			for _, alloc := range removedAllocations {
				klog.Infof("Removing ip %s because is no longer used by service %s/%s", alloc.Address, service.GetNamespace(), service.GetName())
				persistentips.ReleaseIP(alloc.Pool, service.GetNamespace(), alloc.Address)

				if err := persistentips.DeallocateAddress(alloc); err != nil {
					return nil, err
				}
			}
		}

		if allocErr != nil {
			klog.Errorf("error patching allocation for service %s/%s: %v", service.GetNamespace(), service.GetName(), allocErr)
			return patched, allocErr
		}
		return patched, nil

	}
	err := fmt.Errorf("service %s/%s is expected as Persistent but hasn't external ips", service.GetNamespace(), service.GetName())
	klog.Error(err)
	return nil, err
}

//	-> has externals ips?
//		-> if no:
//			-> check if the allocation has a node and, if necessary, the network interface
//		-> if yes: why? if is ephemeral it shouldn't
func reconcileEphemeralAllocation(service *v1.Service, allocation *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	isEphemeralAllocation := utils.IsEphemeralAllocation(service)

	if isEphemeralAllocation {
		if (len(service.Status.LoadBalancer.Ingress) == 1 && len(allocation.Spec.Allocations) == 1 && service.Status.LoadBalancer.Ingress[0].IP == allocation.Spec.Allocations[0].Address) || len(service.Status.LoadBalancer.Ingress) == 0 {
			patched, allocErr := ephemeralips.CheckAndPatchAllocation(allocation)
			return patched, allocErr
		}
		klog.Warningf("Allocation %s/%s is out of sync with service, deleting and waiting for recreation", allocation.GetNamespace(), allocation.GetName())
		//destruptive situation, better to recreate the allocation
		err := ipallocations.DeleteAllocationByName(allocation.GetNamespace(), allocation.GetName())
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return nil, nil

	}
	err := fmt.Errorf("service %s/%s is expected as Ephemeral but has external ips", service.GetNamespace(), service.GetName())
	klog.Error(err)
	return nil, err
}

// delete and reacreate the allocation
func changeAllocationType(service *v1.Service, allocation *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	currentAllocationType := allocation.Spec.Type
	klog.Infof("Changing allocation of service %s/%s type from %s to %s", allocation.GetNamespace(), allocation.GetName(), currentAllocationType, expectedAllocationType(service))
	err := ipallocations.DeleteAllocationByName(allocation.GetNamespace(), allocation.GetName())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return nil, nil
}

// CreateAllocationForService creates a new IPAllocation according to the given service
func CreateAllocationForService(service *v1.Service) (*loadbalancing_v1alpha1.IPAllocation, error) {
	isPersistentallocation, ips := utils.IsPersistentAllocation(service)
	if isPersistentallocation && len(ips) > 0 {
		allocation, err := persistentips.EnsurePersistentAllocation(service.GetNamespace(), service.GetName(), ips)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return allocation, nil
	}

	allocation, err := ephemeralips.EnsureEphemeralAllocation(service.GetNamespace(), service.GetName())
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return allocation, nil

}

func serviceIsNoLongerALoadBalancer(service *v1.Service, allocation *loadbalancing_v1alpha1.IPAllocation) error {
	// if now is ehemeral we need to remove ingress addresses
	currentAllocationType := allocation.Spec.Type
	if currentAllocationType == loadbalancing_v1alpha1.EphemeralIP {
		err := servicesupdater.RemoveServiceIngressIPs(service.GetNamespace(), service.GetName())
		if err != nil {
			return err
		}
	}

	err := ipallocations.DeleteAllocationByName(allocation.GetNamespace(), allocation.GetName())
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func expectedAllocationType(service *v1.Service) loadbalancing_v1alpha1.IPType {
	isPersistentallocation, ips := utils.IsPersistentAllocation(service)
	if isPersistentallocation && len(ips) > 0 {
		return loadbalancing_v1alpha1.PersistentIP
	}
	return loadbalancing_v1alpha1.EphemeralIP
}

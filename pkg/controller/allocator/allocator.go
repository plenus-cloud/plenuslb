package allocator

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/ephemeralips"
	"plenus.io/plenuslb/pkg/controller/ipallocations"
	"plenus.io/plenuslb/pkg/controller/persistentips"
	"plenus.io/plenuslb/pkg/controller/servicesupdater"
	"plenus.io/plenuslb/pkg/controller/servicewatcher"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// EnsureIPAllocation ensures the ip is really allocated
// if required, it checks on cloud and machine
func EnsureIPAllocation(allocation *loadbalancing_v1alpha1.IPAllocation) error {
	serviceRO, err := servicewatcher.FindService(allocation.GetNamespace(), allocation.GetName())
	// if service does not exists delete allocation
	if err != nil && err == servicewatcher.ErrServiceNotFound || !utils.ServiceIsLoadBalancer(serviceRO) {
		klog.Infof("Service of allocation %s/%s no longer exists or is not a LoadBalancer, deleting allocation", allocation.GetNamespace(), allocation.GetName())
		err := ipallocations.DeleteAllocationByName(allocation.GetNamespace(), allocation.GetName())
		if err != nil {
			klog.Error(err)
		}
		return nil
	}

	if err != nil {
		klog.Error(err)
		return err
	}

	ips := []string{}
	for _, addrAllocation := range allocation.Spec.Allocations {
		address := addrAllocation.Address
		poolName := addrAllocation.Pool
		if allocation.Spec.Type == loadbalancing_v1alpha1.PersistentIP {
			persistentips.EnsureAddressIsNotAvailable(poolName, allocation.GetNamespace(), address)
		}

		ips = append(ips, addrAllocation.Address)

		err = allocateAddress(allocation.Spec.Type, addrAllocation)
		if err != nil {
			if err == utils.ErrNoOperatorNodeAvailable || err == utils.ErrFailedToDialWithOperator {
				klog.Errorf("Failed to allocate address %s of pool %s due the following reason %v. Will be retried", address, poolName, err)
				if _, err := ipallocations.SetAllocationStatusNodeError(allocation, fmt.Errorf("Cluster node %s unreachable", addrAllocation.NodeName)); err != nil {
					klog.Error(err)
				}
			} else {
				klog.Error(err)
				if _, err := ipallocations.SetAllocationStatusError(allocation, err); err != nil {
					klog.Error(err)
				}
			}
			return err
		}
	}

	if _, err := ipallocations.SetAllocationStatusSuccess(allocation); err != nil {
		klog.Error(err)
		return err
	}

	if allocation.Spec.Type != loadbalancing_v1alpha1.PersistentIP {
		if err := updateServiceIngressWithIps(allocation.GetNamespace(), allocation.GetName(), ips); err != nil {
			klog.Error(err)
			return err
		}
	}

	return nil
}

func allocateAddress(allocationType loadbalancing_v1alpha1.IPType, addressAllocation *loadbalancing_v1alpha1.IPAllocationAddresses) error {
	// I retrieve here the pool in order to be able to handle some hot-change to the ippool definition
	// I thought this because maybe the allocation is failing due a wrong ippool configuration
	// in this way I can handle that
	if allocationType == loadbalancing_v1alpha1.PersistentIP {
		if err := persistentips.AllocateAddress(addressAllocation); err != nil {
			klog.Error(err)
			return err
		}
		return nil
	}

	if allocationType == loadbalancing_v1alpha1.EphemeralIP {
		if err := ephemeralips.AllocateAddress(addressAllocation); err != nil {
			klog.Error(err)
			return err
		}
		return nil
	}

	err := fmt.Errorf("Unknown allocation type %s", allocationType)
	klog.Error(err)
	return err

}

func updateServiceIngressWithIps(serviceNamespace, serviceName string, ips []string) error {
	err := servicesupdater.UpdateServiceIngressWithIps(serviceNamespace, serviceName, ips)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("Service of allocation %s/%s not found, deleting allocation", serviceNamespace, serviceName)
			err := ipallocations.DeleteAllocationByName(serviceNamespace, serviceName)
			if err != nil {
				klog.Error(err)
			}
			return err
		}
	}
	return err
}

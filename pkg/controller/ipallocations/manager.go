package ipallocations

import (
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/clients"
)

// ErrMoreThanOneEphemeralIP is returned when is requested the creation/pdate of ipallocation with more than one address
var ErrMoreThanOneEphemeralIP = errors.New("Ephemeral allocation can contains at most one address")

// CreateAllocation creates a new IPAllocation object
func CreateAllocation(namespace, name string, allocationType loadbalancing_v1alpha1.IPType, allocations []*loadbalancing_v1alpha1.IPAllocationAddresses) (*loadbalancing_v1alpha1.IPAllocation, error) {
	klog.Infof("Creating allocation %s/%s", namespace, name)
	if allocationType == loadbalancing_v1alpha1.EphemeralIP && len(allocations) > 1 {
		return nil, ErrMoreThanOneEphemeralIP
	}

	ipAllocation := &loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "plenuslb",
			},
		},
		Spec: loadbalancing_v1alpha1.IPAllocationSpec{
			Allocations: allocations,
			Type:        allocationType,
		},
	}

	var createdAllocation *loadbalancing_v1alpha1.IPAllocation
	// More Info:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		allocation, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(namespace).Create(ipAllocation)
		if err != nil {
			klog.Error(err)
			return err
		}
		createdAllocation = allocation
		return nil
	})
	if retryErr != nil {
		klog.Error(retryErr)
		return nil, retryErr
	}
	klog.Infof("Allocation created %s/%s", namespace, name)
	return createdAllocation, nil
}

// UpdateAllocation updates an existing allocation with the given object
func UpdateAllocation(allocationRO *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	var updatedAllocation *loadbalancing_v1alpha1.IPAllocation
	// More Info:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		allocation, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocationRO.GetNamespace()).Update(allocationRO)
		if err != nil {
			klog.Error(err)
			return err
		}
		updatedAllocation = allocation
		return nil
	})
	if retryErr != nil {
		klog.Error(retryErr)
		return nil, retryErr
	}
	klog.Infof("Allocation updated %s/%s", allocationRO.GetNamespace(), allocationRO.GetName())
	return updatedAllocation, nil
}

// SetAllocationStatusSuccess changes the IPAllocation status to success
func SetAllocationStatusSuccess(allocationRO *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation := allocationRO.DeepCopy()
	allocation.Status.State = loadbalancing_v1alpha1.AllocationStatusSuccess
	allocation.Status.Message = "Allocated"
	result, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocation.GetNamespace()).UpdateStatus(allocation)
	if err != nil {
		klog.Error(err)
	}
	return result, err
}

// SetAllocationStatusError changes the IPAllocation status to error
// this status means that the allocation was not possible, the reson is in the .status..message field
func SetAllocationStatusError(allocationRO *loadbalancing_v1alpha1.IPAllocation, reason error) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation := allocationRO.DeepCopy()
	allocation.Status.State = loadbalancing_v1alpha1.AllocationStatusError
	allocation.Status.Message = reason.Error()
	klog.Infof(
		"Updating allocation %s/%s status to %s for the follwing reason: %s",
		allocation.GetNamespace(),
		allocation.GetName(),
		allocation.Status.State,
		allocation.Status.Message,
	)

	result, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocation.GetNamespace()).UpdateStatus(allocation)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// SetAllocationStatusNodeError changes the IPAllocation status to node_error.
// this status means that at least one of the addresses in this allocation is on a node in error state
func SetAllocationStatusNodeError(allocationRO *loadbalancing_v1alpha1.IPAllocation, reason error) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation := allocationRO.DeepCopy()
	allocation.Status.State = loadbalancing_v1alpha1.AllocationStatusNodeError
	allocation.Status.Message = reason.Error()
	klog.Infof(
		"Updating allocation %s/%s status to %s for the follwing reason: %s",
		allocation.GetNamespace(),
		allocation.GetName(),
		allocation.Status.State,
		allocation.Status.Message,
	)

	result, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocation.GetNamespace()).UpdateStatus(allocation)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// SetAllocationStatusFailed changes the IPAllocation status to failed.
func SetAllocationStatusFailed(allocationRO *loadbalancing_v1alpha1.IPAllocation, reason error) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation := allocationRO.DeepCopy()
	allocation.Status.State = loadbalancing_v1alpha1.AllocationStatusFailed
	allocation.Status.Message = reason.Error()
	klog.Infof(
		"Updating allocation %s/%s status to %s for the follwing reason: %s",
		allocation.GetNamespace(),
		allocation.GetName(),
		allocation.Status.State,
		allocation.Status.Message,
	)

	result, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocation.GetNamespace()).UpdateStatus(allocation)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// SetAllocationStatusAddrDeleted changes the IPAllocation status to address_deleted_from_pool.
// this status means that at least one of the addresses in this allocation has been removed from his pool
func SetAllocationStatusAddrDeleted(allocationRO *loadbalancing_v1alpha1.IPAllocation, addr string) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation := allocationRO.DeepCopy()
	allocation.Status.State = loadbalancing_v1alpha1.AllocationStatusAddrDeleted
	allocation.Status.Message = fmt.Sprintf("Address %s removed from pool", addr)
	klog.Infof(
		"Updating allocation %s/%s status to %s for the follwing reason: %s",
		allocation.GetNamespace(),
		allocation.GetName(),
		allocation.Status.State,
		allocation.Status.Message,
	)

	result, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocation.GetNamespace()).UpdateStatus(allocation)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// SetAllocationStatusPending changes the IPAllocation status to pending.
// this status means that the allocation has been successfully created but the addresses haven't been allocated to cloud and/or host.
// the allocator will take care the objects in this state
func SetAllocationStatusPending(allocationRO *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation := allocationRO.DeepCopy()
	allocation.Status.State = loadbalancing_v1alpha1.AllocationStatusPending
	allocation.Status.Message = "Waiting for allocator"
	klog.Infof(
		"Updating allocation %s/%s status to %s for the follwing reason: %s",
		allocation.GetNamespace(),
		allocation.GetName(),
		allocation.Status.State,
		allocation.Status.Message,
	)

	result, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocation.GetNamespace()).UpdateStatus(allocation)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// RemoveAddressFromAllocation removes a specific address from the given allocation ad sets the allocation status in address_deleted_from_pool
func RemoveAddressFromAllocation(allocationRO *loadbalancing_v1alpha1.IPAllocation, removedAddress string) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation := allocationRO.DeepCopy()

	addrAllocations := []*loadbalancing_v1alpha1.IPAllocationAddresses{}
	for _, addrAllocation := range allocationRO.Spec.Allocations {
		if addrAllocation.Address != removedAddress {
			addrAllocations = append(addrAllocations, addrAllocation)
		}
	}

	if len(addrAllocations) != len(allocation.Spec.Allocations) {
		allocation.Spec.Allocations = addrAllocations

		updated, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocation.GetNamespace()).Update(allocation)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		result, err := SetAllocationStatusAddrDeleted(updated, removedAddress)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}
	return allocation, nil
}

// DeleteAllocationByName deletes an IPAllocation object
func DeleteAllocationByName(allocationNamespace, allocationName string) error {
	err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocationNamespace).Delete(allocationName, &metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("Failed to delete ip allocation %s/%s", allocationNamespace, allocationName)
		klog.Error(err)
		return err
	}
	return nil
}

// FindAllocation find an allocation. returns (nill, nil) if not found
func FindAllocation(allocationNamespace, allocationName string) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(allocationNamespace).Get(allocationName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Errorf("Failed to delete ip allocation %s/%s", allocationNamespace, allocationName)
			klog.Error(err)
			return nil, err
		}
		return nil, nil
	}
	return allocation, nil
}

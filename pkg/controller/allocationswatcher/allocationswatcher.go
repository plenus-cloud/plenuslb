package allocationswatcher

import (
	"errors"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	allocationreconciler "plenus.io/plenuslb/pkg/controller/allocationReconciler"
	allocationslock "plenus.io/plenuslb/pkg/controller/allocationsLock"
	"plenus.io/plenuslb/pkg/controller/allocator"
	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/ephemeralips"
	"plenus.io/plenuslb/pkg/controller/events"
	"plenus.io/plenuslb/pkg/controller/ipallocations"
	operatorspeaker "plenus.io/plenuslb/pkg/controller/operatorSpeaker"
	"plenus.io/plenuslb/pkg/controller/persistentips"
	"plenus.io/plenuslb/pkg/controller/servicesupdater"
	"plenus.io/plenuslb/pkg/controller/servicewatcher"
	"plenus.io/plenuslb/pkg/controller/utils"
	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
)

var (
	allocationStore cache.Store
	// AllocationController is the allocation cache controller
	AllocationController cache.Controller
)

// ErrAllocationNotFound returned when the requested allocation does not exists
var ErrAllocationNotFound = errors.New("Allocation not found")

// Init performs all the sturtup tasks, such as the subscription on the events channel
func Init() {
	buildAllocationsWhatcher()
	events.RegisterOnPersistentPoolModifiedFunc(persistentPoolModified)
	events.RegisterOnPersistentPoolDeletedFunc(persistentPoolRemoved)
	events.RegisterOnEphemeralPoolModifiedFunc(ephemeralPoolModified)
	events.RegisterOnEphemeralPoolDeletedFunc(ephemeralPoolRemoved)
	events.RegisterOnOperatorNodeLostFunc(operatorNodeLost)
	events.RegisterOnNewOperatorNodeFunc(newOperatorNode)
}

var getControllerSourceWatchList = func() cache.ListerWatcher {
	return cache.NewListWatchFromClient(
		clients.GetPlenuslbClient().LoadbalancingV1alpha1().RESTClient(),
		string(loadbalancing_v1alpha1.IPAllocationCRDPlural),
		v1.NamespaceAll,
		fields.Everything(),
	)
}

// buildAllocationsWhatcher build the controller anc the store for IPAllocations
func buildAllocationsWhatcher() {
	store, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		getControllerSourceWatchList(),
		&loadbalancing_v1alpha1.IPAllocation{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				klog.Infof("Allocation %s added in namespace %s", allocation.GetName(), allocation.GetNamespace())
				allocationCreated(allocation)
			},
			DeleteFunc: func(obj interface{}) {
				allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				klog.Infof("Allocation %s deleted in namespace %s", allocation.GetName(), allocation.GetNamespace())

				allocationDeleted(allocation)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				allocation, ok := newObj.(*loadbalancing_v1alpha1.IPAllocation)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(newObj))
					return
				}
				klog.Infof("Allocation %s modified from namespace %s", allocation.GetName(), allocation.GetNamespace())

				allocationsChanged(allocation)
			},
		},
	)

	allocationStore = store
	AllocationController = controller
}

// WatchAllocations watch the IPAllocations objects
func WatchAllocations(stop chan struct{}) {
	go AllocationController.Run(stop)
}

func allocationCreated(allocation *loadbalancing_v1alpha1.IPAllocation) {
	switch allocation.Status.State {
	case "":
		break
	case loadbalancing_v1alpha1.AllocationStatusSuccess:
		go processAllocationStatusSuccess(allocation)
		break
	case loadbalancing_v1alpha1.AllocationStatusPending:
		go processAllocationStatusPending(allocation)
		break
	case loadbalancing_v1alpha1.AllocationStatusNodeError:
		go processAllocationNodeError(allocation)
		break
	case loadbalancing_v1alpha1.AllocationStatusError:
		go processAllocationStatusError(allocation)
	case loadbalancing_v1alpha1.AllocationStatusFailed:
		klog.Warningf("Allocation %s/%s is now in failed state. A human action is requied", allocation.GetNamespace(), allocation.GetName())
	default:
		klog.Errorf("State %s not implemented", allocation.Status.State)
	}
}

func allocationsChanged(allocation *loadbalancing_v1alpha1.IPAllocation) {
	switch allocation.Status.State {
	case "":
		break
	case loadbalancing_v1alpha1.AllocationStatusSuccess:
		go processAllocationStatusSuccess(allocation)
		break
	case loadbalancing_v1alpha1.AllocationStatusPending:
		go processAllocationStatusPending(allocation)
		break
	case loadbalancing_v1alpha1.AllocationStatusNodeError:
		go processAllocationNodeError(allocation)
		break
	case loadbalancing_v1alpha1.AllocationStatusError:
		go processAllocationStatusError(allocation)
	case loadbalancing_v1alpha1.AllocationStatusFailed:
		klog.Warningf("Allocation %s/%s is now in failed state. A human action is requied", allocation.GetNamespace(), allocation.GetName())
	default:
		klog.Errorf("State %s not implemented", allocation.Status.State)
	}
}

func processAllocationStatusError(allocationRO *loadbalancing_v1alpha1.IPAllocation) error {
	if allocationslock.IsErrorAllocationAlreadyProcessing(allocationRO) {
		klog.Infof("Error allocation %s/%s is already processing, not added to backoff", allocationRO.GetNamespace(), allocationRO.GetName())
		return nil
	}
	allocationslock.AddErrorAllocationToProcessingList(allocationRO)
	// retry until timeout
	err := utils.OnErrorForever(utils.ErrorBackoff, func(err error) bool { return true }, func() (err error) {
		if err := allocationslock.AcquireAllocationLock(allocationRO); err != nil {
			return err
		}
		defer allocationslock.RemoveFromLock(allocationRO)
		allocation, err := FindAllocationByName(allocationRO.GetNamespace(), allocationRO.GetName())
		if err != nil {
			return err
		}
		if allocation.Status.State != loadbalancing_v1alpha1.AllocationStatusError {
			klog.Infof("Allocation %s/%s state is no loger error, now is %s", allocation.GetNamespace(), allocation.GetName(), allocation.Status.State)
			allocationslock.RemoveErrorAllocationFromProcessingList(allocationRO)
			return nil
		}
		klog.Infof("Processing error allocation %s/%s (Error: %s)", allocation.GetNamespace(), allocation.GetName(), allocation.Status.Message)

		service, err := servicewatcher.FindService(allocation.GetNamespace(), allocation.GetName())
		if err != nil {
			klog.Error(err)
			return err
		}
		reconciledAllocation, err := allocationreconciler.ReconcileAllocation(service, allocation)
		if err != nil {
			klog.Error(err)
			return err
		}

		// ensure the IP allocation only if the IPAllocation object is equal before reconciliation
		// if are different, the update event will take care of this
		if reflect.DeepEqual(reconciledAllocation, allocation) {
			err := allocator.EnsureIPAllocation(reconciledAllocation)
			if err != nil {
				klog.Error(err)
				return err
			}
		}

		return errors.New(reconciledAllocation.Status.Message)
	})
	if err != nil {
		allocationslock.RemoveErrorAllocationFromProcessingList(allocationRO)
		allocation, findErr := FindAllocationByName(allocationRO.GetNamespace(), allocationRO.GetName())
		if findErr != nil {
			return findErr
		}
		klog.Error(err)
		_, err = ipallocations.SetAllocationStatusFailed(allocation, err)
		return err
	}
	return nil
}

func processAllocationStatusSuccess(allocationRO *loadbalancing_v1alpha1.IPAllocation) error {
	if err := allocationslock.AcquireAllocationLock(allocationRO); err != nil {
		return err
	}
	defer allocationslock.RemoveFromLock(allocationRO)

	allocation, err := FindAllocationByName(allocationRO.GetNamespace(), allocationRO.GetName())
	if err != nil {
		return err
	}

	if allocation.Status.State != loadbalancing_v1alpha1.AllocationStatusSuccess {
		klog.Infof("Allocation %s/%s state is no loger success, now is %s", allocation.GetNamespace(), allocation.GetName(), allocation.Status.State)
		return nil
	}
	klog.Infof("Processing success allocation %s/%s", allocation.GetNamespace(), allocation.GetName())

	service, err := servicewatcher.FindService(allocation.GetNamespace(), allocation.GetName())
	if err != nil {
		klog.Error(err)
		return err
	}
	reconciledAllocation, err := allocationreconciler.ReconcileAllocation(service, allocation)
	if err != nil {
		klog.Error(err)
		return err
	}

	// ensure the IP allocation only if the IPAllocation object is equal before reconciliation
	// if are different, the update event will take care of this
	if reflect.DeepEqual(reconciledAllocation, allocation) {
		err := allocator.EnsureIPAllocation(reconciledAllocation)
		if err != nil {
			klog.Error(err)
			return err
		}
	}

	return nil
}

func allocationDeleted(allocation *loadbalancing_v1alpha1.IPAllocation) {
	go func() {
		if err := allocationslock.AcquireAllocationLock(allocation); err != nil {
			return
		}
		defer allocationslock.RemoveFromLock(allocation)
		allocationType := allocation.Spec.Type
		if allocationType == loadbalancing_v1alpha1.PersistentIP {
			persistentips.DeallocateAllocation(allocation)
		} else if allocationType == loadbalancing_v1alpha1.EphemeralIP {
			ephemeralips.DeallocateAddress(allocation)
			servicesupdater.RemoveServiceIngressIPs(allocation.GetNamespace(), allocation.GetName())
		} else {
			err := fmt.Errorf("Unknown allocation type %s", allocationType)
			klog.Error(err)
		}

		service, err := servicewatcher.FindService(allocation.GetNamespace(), allocation.GetName())
		if err == nil && utils.ServiceIsLoadBalancer(service) {
			klog.Infof("Service of deleted allocation %s/%s still exists and is a LoadBalancer, recreating allocation", allocation.GetNamespace(), allocation.GetName())
			allocationreconciler.CreateAllocationForService(service)
		}
	}()
}

// operatorNodeLost move all addresses on a dead node somewhere else
func operatorNodeLost(clusterNodeName string) {
	klog.Infof("Moving all addresses allocated on lost node %s", clusterNodeName)
	allocations := allocationStore.List()

	// delete the allocations on that node, will be recreated somewhere
	for _, obj := range allocations {
		allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
		if !ok {
			err := fmt.Errorf("unexpected type %s", reflect.TypeOf(obj))
			klog.Error(err)
			continue
		}

		for _, addrAlloc := range allocation.Spec.Allocations {
			if addrAlloc.NodeName == clusterNodeName {
				klog.Infof("Allocation %s/%s is on lost node %s, changing allocation status", allocation.GetNamespace(), allocation.GetName(), clusterNodeName)
				if result, err := ipallocations.SetAllocationStatusNodeError(allocation, fmt.Errorf("Cluster node %s lost", clusterNodeName)); err != nil {
					klog.Error(err)
				} else {
					allocation = result
				}
				break
			}
		}
	}
}

// newOperatorNode performs some cleanup on new node
func newOperatorNode(clusterNodeName string) {
	klog.Infof("Cleaning up not required addresses from new node %s", clusterNodeName)
	allocations := allocationStore.List()

	toKeep := []*plenuslbV1Alpha1.AddressInfo{}

	// delete the allocations on that node, will be recreated somewhere
	for _, obj := range allocations {
		allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
		if !ok {
			err := fmt.Errorf("unexpected type %s", reflect.TypeOf(obj))
			klog.Error(err)
			continue
		}

		for _, addrAlloc := range allocation.Spec.Allocations {
			if addrAlloc.NodeName == clusterNodeName {
				klog.Infof("Address %s on interface %s used by allocation %s/%s must be kept", addrAlloc.Address, addrAlloc.NetworkInterface, allocation.GetNamespace(), allocation.GetName())
				toKeep = append(toKeep, &plenuslbV1Alpha1.AddressInfo{
					Address:   addrAlloc.Address,
					Interface: addrAlloc.NetworkInterface,
				})
			}
		}
	}

	operatorspeaker.DoCleanup(clusterNodeName, toKeep)
}

func deallocateDeletedAddresses(addresses []string) {
	allocations := allocationStore.List()

	for _, obj := range allocations {
		allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
		if !ok {
			err := fmt.Errorf("unexpected type %s", reflect.TypeOf(obj))
			klog.Error(err)
			continue
		}

		for _, addrAlloc := range allocation.Spec.Allocations {
			if utils.ContainsString(addresses, addrAlloc.Address) {
				klog.Infof("Address %s has been removed from pool, removing from allocation", addrAlloc.Address)
				if result, err := ipallocations.RemoveAddressFromAllocation(allocation, addrAlloc.Address); err != nil {
					klog.Error(err)
				} else {
					allocation = result
				}
			}

		}
	}
}

func deallocateAddressesOfPool(poolName string) {
	allocations := allocationStore.List()

	for _, obj := range allocations {
		allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
		if !ok {
			err := fmt.Errorf("unexpected type %s", reflect.TypeOf(obj))
			klog.Error(err)
			continue
		}

		for _, addrAlloc := range allocation.Spec.Allocations {
			if addrAlloc.Pool == poolName {
				klog.Infof("Address %s has been removed from pool, removing from allocation %s/%s", addrAlloc.Address, allocation.GetNamespace(), allocation.GetName())
				if result, err := ipallocations.RemoveAddressFromAllocation(allocation, addrAlloc.Address); err != nil {
					klog.Error(err)
				} else {
					allocation = result
				}
			}

		}
	}
}

func persistentPoolModified(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	for _, obj := range allocationStore.List() {
		allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			continue
		}

		if allocation.Spec.Type != loadbalancing_v1alpha1.PersistentIP {
			continue
		}

		for _, addrAllocation := range allocation.Spec.Allocations {
			if addrAllocation.Pool == pool.GetName() {
				if !utils.PoolHasAddress(pool, addrAllocation.Address) {
					klog.Infof("Address %s has been removed from pool %s, deallocating", addrAllocation.Address, pool.GetName())
					if result, err := ipallocations.RemoveAddressFromAllocation(allocation, addrAllocation.Address); err != nil {
						klog.Error(err)
					} else {
						allocation = result
					}
				}
			}
		}
	}

	persistentips.ProcessIPAvailabilityFromCacheList(pool, allocationStore.List())
}

func ephemeralPoolModified(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	// TODO
	// host network changed?
}

func persistentPoolRemoved(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	deallocateDeletedAddresses(pool.Spec.Addresses)
}

func ephemeralPoolRemoved(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	deallocateAddressesOfPool(pool.GetName())
}

func processAllocationNodeError(allocationRO *loadbalancing_v1alpha1.IPAllocation) error {
	if err := allocationslock.AcquireAllocationLock(allocationRO); err != nil {
		return err
	}
	defer allocationslock.RemoveFromLock(allocationRO)

	allocation, err := FindAllocationByName(allocationRO.GetNamespace(), allocationRO.GetName())
	if err != nil {
		return err
	}

	if allocation.Status.State != loadbalancing_v1alpha1.AllocationStatusNodeError {
		return nil
	}

	if allocation.Spec.Type == loadbalancing_v1alpha1.PersistentIP {
		_, err = persistentips.ChangeAllocationNode(allocation)
	} else {
		_, err = ephemeralips.ChangeAllocationNode(allocation)
	}

	return err
}

func processAllocationStatusPending(allocationRO *loadbalancing_v1alpha1.IPAllocation) error {
	if err := allocationslock.AcquireAllocationLock(allocationRO); err != nil {
		return err
	}
	defer allocationslock.RemoveFromLock(allocationRO)

	allocation, err := FindAllocationByName(allocationRO.GetNamespace(), allocationRO.GetName())
	if err != nil {
		return err
	}

	if allocation.Status.State != loadbalancing_v1alpha1.AllocationStatusPending {
		return nil
	}

	err = allocator.EnsureIPAllocation(allocation)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// FindAllocationByName returns the allocation given the allocation name and namespace
func FindAllocationByName(namespace, name string) (*loadbalancing_v1alpha1.IPAllocation, error) {
	/* 	if AllocationController != nil && AllocationController.HasSynced() {
		for _, obj := range allocationStore.List() {
			allocation, ok := obj.(*loadbalancing_v1alpha1.IPAllocation)
			if !ok {
				klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
				continue
			}
			if allocation.GetNamespace() == namespace && allocation.GetName() == name {
				return allocation, nil
			}
		}
		return nil, ErrAllocationNotFound
	} */

	allocation, err := clients.GetPlenuslbClient().LoadbalancingV1alpha1().IPAllocations(namespace).Get(name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		klog.Warningf("Allocation %s/%s not found", namespace, name)
		return nil, ErrAllocationNotFound
	} else if err != nil {
		return nil, err
	}

	return allocation, nil
}

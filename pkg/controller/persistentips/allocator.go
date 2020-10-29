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
	"fmt"
	"math/rand"
	"reflect"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/ipallocations"
	"plenus.io/plenuslb/pkg/controller/operator"
	operatorspeaker "plenus.io/plenuslb/pkg/controller/operatorSpeaker"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// EnsurePersistentAllocation makes sure the service has the ip allocation
// This function is called by reconciliation function
func EnsurePersistentAllocation(serviceNamespace, serviceName string, ips []string) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation, err := ipallocations.FindAllocation(serviceNamespace, serviceName)
	if err != nil {
		return nil, err
	} else if allocation != nil {
		return allocation, nil
	}

	allocation, allocationErr, err := createPersistentAllocation(serviceNamespace, serviceName, ips)
	if err != nil {
		return nil, err
	}

	if allocationErr != nil {
		allocation, err := ipallocations.SetAllocationStatusError(allocation, allocationErr)
		return allocation, err
	}
	allocation, err = ipallocations.SetAllocationStatusPending(allocation)
	return allocation, err
}

// CheckAndPatchAllocation checks the current persistent allocation and, if required, patches it
// in order to syncronize the allocation with the pool, the nodes and the cloud
// This function is called by reconciliation function
func CheckAndPatchAllocation(
	serviceNamespace,
	serviceName string,
	ips []string,
	currentAllocationRO *loadbalancing_v1alpha1.IPAllocation,
) (
	*loadbalancing_v1alpha1.IPAllocation,
	[]*loadbalancing_v1alpha1.IPAllocationAddresses,
	error,
	error,
) {
	actualAllocation := currentAllocationRO.DeepCopy()
	allocations, allocationErr := buildAllocations(serviceNamespace, actualAllocation, ips)

	removedAllocations := []*loadbalancing_v1alpha1.IPAllocationAddresses{}
	for _, alloc := range actualAllocation.Spec.Allocations {
		if contains, _ := utils.ContainsAddress(allocations, alloc.Address); !contains {
			ReleaseIP(alloc.Pool, serviceNamespace, alloc.Address)
			removedAllocations = append(
				removedAllocations,
				alloc.DeepCopy(),
			)
		}
	}
	actualAllocation.Spec.Allocations = allocations
	if allocationErr == nil && reflect.DeepEqual(actualAllocation, currentAllocationRO) {
		klog.Infof("Nothing to do on allocation %s/%s", currentAllocationRO.GetNamespace(), currentAllocationRO.GetName())
		return actualAllocation, removedAllocations, allocationErr, nil
	}

	res, err := ipallocations.UpdateAllocation(actualAllocation)
	if err != nil {
		klog.Error(err)
		return nil, nil, allocationErr, err
	}

	if allocationErr != nil {
		klog.Error(allocationErr)
		res, err := ipallocations.SetAllocationStatusError(res, allocationErr)
		if err != nil {
			return nil, removedAllocations, allocationErr, err
		}
		return res, removedAllocations, allocationErr, err
	}
	res, err = ipallocations.SetAllocationStatusPending(res)
	if err != nil {
		klog.Error(err)
		return nil, removedAllocations, allocationErr, err
	}
	return res, removedAllocations, allocationErr, err

}

func addOperatorNodeToAllocation(pool *loadbalancing_v1alpha1.PersistentIPPool, allocationRO *loadbalancing_v1alpha1.IPAllocationAddresses) (*loadbalancing_v1alpha1.IPAllocationAddresses, error) {
	allocation := allocationRO.DeepCopy()
	var allocationErr error
	hasHostNetworkOption := utils.PersistentPoolHasHostNetworkOption(pool)
	hasCloudIntegrationOption, cloudProvider := utils.PersistentPoolHasCloudIntegrationOption(pool)

	if hasHostNetworkOption && (allocation.NodeName == "" || allocation.NetworkInterface == "" || allocation.NetworkInterface != pool.Spec.Options.HostNetworkInterface.InterfaceName) {
		// pick a node
		operatorNode, err := operator.GetRandomOperatorNode()
		if err != nil {
			klog.Error(err)
			allocationErr = err
		} else {
			allocation.NodeName = operatorNode.NodeName
		}
		// add address to node
		allocation.NetworkInterface = pool.Spec.Options.HostNetworkInterface.InterfaceName

	} else if hasCloudIntegrationOption && allocation.NodeName == "" {
		clusterNode, err := getRandomNode()
		if err != nil {
			klog.Error(err)
			allocationErr = err
		} else {
			allocation.NodeName = clusterNode.GetName()
		}
	}

	allocation.CloudProvider = cloudProvider

	return allocation, allocationErr
}

func buildAllocations(serviceNamespace string, actualAllocationRO *loadbalancing_v1alpha1.IPAllocation, ips []string) ([]*loadbalancing_v1alpha1.IPAllocationAddresses, error) {
	var allocationErr error
	allocations := []*loadbalancing_v1alpha1.IPAllocationAddresses{}

	for _, ip := range ips {
		allocation := &loadbalancing_v1alpha1.IPAllocationAddresses{
			Address: ip,
		}
		var pool *loadbalancing_v1alpha1.PersistentIPPool
		var contains bool
		var ready *loadbalancing_v1alpha1.IPAllocationAddresses
		if actualAllocationRO != nil {
			contains, ready = utils.ContainsAddress(actualAllocationRO.Spec.Allocations, ip)
		}
		if contains {
			pool = GetPoolOfAddress(serviceNamespace, ready.Address)
			allocation = ready.DeepCopy()
			if pool == nil {
				allocationErr := fmt.Errorf("Pool for address %s of allocation %s/%s not found", ready.Address, actualAllocationRO.GetNamespace(), actualAllocationRO.GetName())
				klog.Error(allocationErr)
			}
		} else {
			var err error
			pool, err = UseIP(serviceNamespace, ip)
			if err != nil {
				allocationErr = err
				klog.Error(err)
			} else {
				allocation.Pool = pool.GetName()
			}
		}

		if pool != nil {
			var err error
			allocation, err = addOperatorNodeToAllocation(pool, allocation)
			if err != nil {
				allocationErr = err
			}
			allocations = append(allocations, allocation)
		}
	}

	return allocations, allocationErr
}

func createPersistentAllocation(serviceNamespace, serviceName string, ips []string) (*loadbalancing_v1alpha1.IPAllocation, error, error) {
	allocations, allocationErr := buildAllocations(serviceNamespace, nil, ips)
	allocation, err := ipallocations.CreateAllocation(serviceNamespace, serviceName, loadbalancing_v1alpha1.PersistentIP, allocations)
	if err != nil {
		for _, alloc := range allocations {
			ReleaseIP(alloc.Pool, serviceNamespace, alloc.Address)
		}
		klog.Error(err)
	}
	return allocation, allocationErr, err
}

// ChangeAllocationNode moves the ip allocation from one node to one other
// if the pool ha cloud integration option, the change will be propagated to the cloud
// if the pool has the host network option, the change will be propagated to the host machine
func ChangeAllocationNode(allocationRO *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	klog.Infof("Changing node for allocation %s/%s", allocationRO.GetNamespace(), allocationRO.GetName())
	var allocationErr error
	allocation := allocationRO.DeepCopy()
	for _, allocation := range allocation.Spec.Allocations {
		pool := SearchPoolByName(allocation.Pool)
		if pool == nil {
			e := ErrPoolNotFound
			allocationErr = e
			klog.Error(e)
			break
		}

		nodeName := ""
		netInterface := ""

		hasHostNetworkOption := utils.PersistentPoolHasHostNetworkOption(pool)
		hasCloudIntegrationOption, _ := utils.PersistentPoolHasCloudIntegrationOption(pool)
		if hasHostNetworkOption {
			// pick a node
			operatorNode, err := operator.GetRandomOperatorNode()
			if err != nil {
				klog.Error(err)
				allocationErr = err
			} else {
				nodeName = operatorNode.NodeName
			}
			// add address to node
			netInterface = pool.Spec.Options.HostNetworkInterface.InterfaceName

		} else if hasCloudIntegrationOption {
			clusterNode, err := getRandomNode()
			if err != nil {
				klog.Error(err)
				allocationErr = err
			} else {
				nodeName = clusterNode.GetName()
			}
		}

		allocation.NetworkInterface = netInterface
		allocation.NodeName = nodeName
		klog.Infof("Allocation %s/%s is now on interface %s of node %s", allocationRO.GetNamespace(), allocationRO.GetName(), netInterface, nodeName)
	}

	allocation, err := ipallocations.UpdateAllocation(allocation)
	if err != nil {
		klog.Error(err)
		allocationErr = err
	}

	if allocationErr != nil {
		result, err := ipallocations.SetAllocationStatusError(allocation, allocationErr)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return result, nil
	}

	result, err := ipallocations.SetAllocationStatusPending(allocation)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil

}

func getNetworkNodesList() (*v1.NodeList, error) {
	nodeList, err := clients.GetK8sClient().CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return nodeList, nil
}

func getRandomNode() (*v1.Node, error) {
	nodeList, err := getNetworkNodesList()
	if err != nil {
		return nil, err
	}

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator
	i := r.Intn(len(nodeList.Items))

	node := nodeList.Items[i] // Truncc.poolste slice.

	return &node, nil
}

// AllocateAddress allocates an ip address declared on a pool
// if the pool has a cloud integration, the ip wil be bought through the cloud
// if the pool has the  host network option, the ip will be added to the host machine
func AllocateAddress(addressAllocation *loadbalancing_v1alpha1.IPAllocationAddresses) error {
	pool := SearchPoolByName(addressAllocation.Pool)
	if pool == nil {
		klog.Errorf("Cannot find pool %s", addressAllocation.Pool)
		return ErrPoolNotFound
	}
	hasHostNetworkOption := utils.PersistentPoolHasHostNetworkOption(pool)
	hasCloudIntegration, cloudName := utils.PersistentPoolHasCloudIntegrationOption(pool)
	if hasCloudIntegration || hasHostNetworkOption {
		clusterNodeName := addressAllocation.NodeName
		if hasHostNetworkOption {
			netInterface := pool.Spec.Options.HostNetworkInterface.InterfaceName
			klog.Infof("Ensuring allocation of address %s on interface %s of node %s", addressAllocation.Address, netInterface, addressAllocation.NodeName)
			// add address to node
			if err := operatorspeaker.EnsureIPAllocationOnNode(addressAllocation.NodeName, netInterface, addressAllocation.Address); err != nil {
				return err
			}
		}

		if hasCloudIntegration {
			klog.Infof("Ensuring allocation of address %s on %s cloud node %s", addressAllocation.Address, cloudName, addressAllocation.NodeName)
			err := allocateAddressOnCloud(pool, addressAllocation.Address, clusterNodeName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeallocateAddress deallocates an ip
// if the pool has the cloud integration, the ip wil be deallocated on the cloud
// if the pool has the host network option, the ip will be removed from the host machine
func DeallocateAddress(addressAllocation *loadbalancing_v1alpha1.IPAllocationAddresses) error {
	pool := SearchPoolByName(addressAllocation.Pool)
	if pool == nil {
		klog.Errorf("Cannot find pool %s", addressAllocation.Pool)
		return ErrPoolNotFound
	}
	hasHostNetworkOption := utils.PersistentPoolHasHostNetworkOption(pool)
	hasCloudIntegration, cloudName := utils.PersistentPoolHasCloudIntegrationOption(pool)
	if hasCloudIntegration || hasHostNetworkOption {
		if hasHostNetworkOption {
			netInterface := pool.Spec.Options.HostNetworkInterface.InterfaceName
			klog.Infof("Ensuring deallocation of address %s on interface %s of node %s", addressAllocation.Address, netInterface, addressAllocation.NodeName)
			// add address to node
			if err := operatorspeaker.RemoveAddressFromNode(addressAllocation.NodeName, netInterface, addressAllocation.Address); err != nil {
				return err
			}
		}

		if hasCloudIntegration {
			klog.Infof("Ensuring deallocation of address %s on %s cloud node %s", addressAllocation.Address, cloudName, addressAllocation.NodeName)
			err := deallocateAddressFromCloud(pool, addressAllocation.Address)
			if err != nil {
				klog.Error(err)
				return err
			}
		}
	}
	return nil
}

func allocateAddressOnCloud(pool *loadbalancing_v1alpha1.PersistentIPPool, address, nodeName string) error {
	if pool.Spec.CloudIntegration == nil {
		return nil
	}

	if ci := cloudsIntegration.GetCloudAPI(pool.Spec.CloudIntegration); ci != nil {
		return ci.AssignIPToServer(address, nodeName)
	}
	return nil
}

func deallocateAddressFromCloud(pool *loadbalancing_v1alpha1.PersistentIPPool, address string) error {
	if pool.Spec.CloudIntegration == nil {
		return nil
	}

	if ci := cloudsIntegration.GetCloudAPI(pool.Spec.CloudIntegration); ci != nil {
		return ci.UnassignIP(address)
	}
	return nil

}

// DeallocateAllocation removes ip from nodes, releases the address to the pool and deallcates the address from the cloud
func DeallocateAllocation(allocation *loadbalancing_v1alpha1.IPAllocation) {
	klog.Infof("Deallocating persistent addresses of allocation %s/%s", allocation.GetNamespace(), allocation.GetName())
	for _, addrAllocation := range allocation.Spec.Allocations {
		if addrAllocation.NetworkInterface != "" {
			netInterface := addrAllocation.NetworkInterface
			operatorspeaker.RemoveAddressFromNode(addrAllocation.NodeName, netInterface, addrAllocation.Address)
		}

		klog.Infof("Releasing address %s to pool %s", addrAllocation.Address, addrAllocation.Pool)
		ReleaseIP(addrAllocation.Pool, allocation.GetNamespace(), addrAllocation.Address)

		pool := SearchPoolByName(addrAllocation.Pool)
		if pool != nil {
			if err := deallocateAddressFromCloud(pool, addrAllocation.Address); err != nil {
				klog.Error(err)
			}
		}
	}
}

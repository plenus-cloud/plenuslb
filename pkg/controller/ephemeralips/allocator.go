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
	"fmt"
	"math/rand"
	"net"
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

// EnsureEphemeralAllocation makes sure the service has the ip allocation
// This function is called by reconciliation function
func EnsureEphemeralAllocation(serviceNamespace, serviceName string) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation, err := ipallocations.FindAllocation(serviceNamespace, serviceName)
	if err != nil {
		return nil, err
	} else if allocation != nil {
		return allocation, nil
	}

	allocation, allocationErr, err := createEphemeralAllocation(serviceNamespace, serviceName)
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

// CheckAndPatchAllocation checks the current ephemeral allocation and, if required, patches it
// in order to syncronize the allocation with the pool, the nodes and the cloud
// This function is called by reconciliation function
func CheckAndPatchAllocation(allocationRO *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
	allocation := allocationRO.DeepCopy()
	pool := getPoolForService(allocationRO.GetNamespace())
	if pool == nil {
		klog.Errorf("Ephemeral pool for service %s/%s not found", allocationRO.GetNamespace(), allocationRO.GetName())
		return allocation, ErrPoolNotFound
	}

	var allocationErr error

	hasHostNetworkOption := utils.EphemeralPoolHasHostNetworkOption(pool)
	hasCloudIntegrationOption, cloudProvider := utils.EphemeralPoolHasCloudIntegrationOption(pool)
	addressAllocation := allocation.Spec.Allocations[0]
	if hasHostNetworkOption && (addressAllocation.NodeName == "" || addressAllocation.NetworkInterface == "" || addressAllocation.NetworkInterface != pool.Spec.Options.HostNetworkInterface.InterfaceName) {
		// pick a node
		operatorNode, err := operator.GetRandomOperatorNode()
		if err != nil {
			klog.Error(err)
			allocationErr = err
		} else {
			addressAllocation.NodeName = operatorNode.NodeName
		}
		// add address to node
		addressAllocation.NetworkInterface = pool.Spec.Options.HostNetworkInterface.InterfaceName
	} else if hasCloudIntegrationOption && (addressAllocation.NodeName == "" || addressAllocation.CloudProvider != cloudProvider) {
		clusterNode, err := getRandomNode()
		if err != nil {
			klog.Error(err)
			allocationErr = err
		} else {
			addressAllocation.NodeName = clusterNode.GetName()
		}
		addressAllocation.CloudProvider = cloudProvider
	}

	// if ip not valid set error
	addr := net.ParseIP(addressAllocation.Address)
	if addr == nil {
		allocationErr = fmt.Errorf("address '%s' of allocation %s/%s is not valid", addressAllocation.Address, allocationRO.GetNamespace(), allocationRO.GetName())
		if addressAllocation.NodeName != "" {
			klog.Info(allocationErr)
			klog.Infof("Getting new ephemeral address for allocation %s/%s", allocationRO.GetNamespace(), allocationRO.GetName())
			clusterName := utils.GetClusterName()
			ip, err := getAndAssignAddressOnCloud(pool, fmt.Sprintf("plenuslb-ephemeral-%s-%s-%s", clusterName, allocationRO.GetNamespace(), allocationRO.GetName()), addressAllocation.NodeName)
			if err != nil {
				klog.Error(err)
				return allocationRO, err
			}
			addressAllocation.Address = ip
		} else {
			klog.Error(allocationErr)
			return allocationRO, allocationErr
		}
	}

	if reflect.DeepEqual(allocationRO, allocation) {
		klog.Infof("Nothing to do on allocation %s/%s", allocationRO.GetNamespace(), allocationRO.GetName())
		return allocationRO, allocationErr
	}

	var err error
	allocation, err = ipallocations.UpdateAllocation(allocation)
	if err != nil {
		klog.Error(err)
		return allocation, err
	}

	if allocationErr != nil {
		return ipallocations.SetAllocationStatusError(allocation, allocationErr)
	}

	return ipallocations.SetAllocationStatusPending(allocation)

}

func buildAllocations(serviceNamespace, serviceName string) ([]*loadbalancing_v1alpha1.IPAllocationAddresses, error) {
	var allocationErr error
	allocations := []*loadbalancing_v1alpha1.IPAllocationAddresses{}
	pool := getPoolForService(serviceNamespace)
	if pool == nil {
		klog.Errorf("Ephemeral pool for service %s/%s not found", serviceNamespace, serviceName)
		allocationErr = ErrPoolNotFound
	}

	if pool != nil {
		nodeName := ""
		netInterface := ""

		hasHostNetworkOption := utils.EphemeralPoolHasHostNetworkOption(pool)
		hasCloudIntegrationOption, cloudProvider := utils.EphemeralPoolHasCloudIntegrationOption(pool)
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
				allocationErr = utils.ErrFailedToDialWithOperator
			} else {
				nodeName = clusterNode.GetName()
			}
		}

		clusterName := utils.GetClusterName()
		ip, err := getAndAssignAddressOnCloud(pool, fmt.Sprintf("plenuslb-ephemeral-%s-%s-%s", clusterName, serviceNamespace, serviceName), nodeName)
		if err != nil {
			klog.Error(err)
			allocationErr = err
		}

		allocation := loadbalancing_v1alpha1.IPAllocationAddresses{
			Address:          ip,
			NetworkInterface: netInterface,
			NodeName:         nodeName,
			CloudProvider:    cloudProvider,
			Pool:             pool.GetName(),
		}

		allocations = append(allocations, &allocation)
	}
	return allocations, allocationErr
}

func createEphemeralAllocation(serviceNamespace, serviceName string) (*loadbalancing_v1alpha1.IPAllocation, error, error) {
	allocations, allocationErr := buildAllocations(serviceNamespace, serviceName)

	createdAllocation, err := ipallocations.CreateAllocation(serviceNamespace, serviceName, loadbalancing_v1alpha1.EphemeralIP, allocations)
	if err != nil {
		klog.Error(err)
	}
	return createdAllocation, allocationErr, err
}

// ChangeAllocationNode moves the ip allocation from one node to one other
// if the pool ha cloud integration option, the change will be propagated to the cloud
// if the pool has the host network option, the change will be propagated to the host machine
func ChangeAllocationNode(allocationRO *loadbalancing_v1alpha1.IPAllocation) (*loadbalancing_v1alpha1.IPAllocation, error) {
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

		hasHostNetworkOption := utils.EphemeralPoolHasHostNetworkOption(pool)
		hasCloudIntegrationOption, _ := utils.EphemeralPoolHasCloudIntegrationOption(pool)
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

// AllocateAddress allocates a new ip address
// if the pool has a cloud integration, the ip wil be bought through the cloud
// if the pool has the  host network option, the ip will be added to the host machine
func AllocateAddress(addressAllocation *loadbalancing_v1alpha1.IPAllocationAddresses) error {
	pool := SearchPoolByName(addressAllocation.Pool)
	if pool == nil {
		klog.Errorf("Cannot find pool %s", addressAllocation.Pool)
		return ErrPoolNotFound
	}
	hasHostNetworkOption := utils.EphemeralPoolHasHostNetworkOption(pool)
	hasCloudIntegration, cloudName := utils.EphemeralPoolHasCloudIntegrationOption(pool)
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
// if the pool has the cloud integration, the ip wil be released to the cloud
// if the pool has the host network option, the ip will be removed from the host machine
func DeallocateAddress(allocation *loadbalancing_v1alpha1.IPAllocation) {
	klog.Infof("Deallocating ephemeral addresses of allocation %s/%s", allocation.GetNamespace(), allocation.GetName())
	for _, addrAllocation := range allocation.Spec.Allocations {
		if addrAllocation.NetworkInterface != "" {
			netInterface := addrAllocation.NetworkInterface
			_ = operatorspeaker.RemoveAddressFromNode(addrAllocation.NodeName, netInterface, addrAllocation.Address)
		}

		pool := SearchPoolByName(addrAllocation.Pool)
		if pool != nil {
			if err := deleteAddressFromCloud(pool, addrAllocation.Address); err != nil {
				klog.Error(err)
			}
		} else {
			klog.Errorf("Cannot find ephemeral pool by name %s", addrAllocation.Pool)
		}
	}
}

func getAndAssignAddressOnCloud(pool *loadbalancing_v1alpha1.EphemeralIPPool, ipName, nodeName string) (string, error) {
	if pool.Spec.CloudIntegration == nil {
		return "", nil
	}

	if ci := cloudsIntegration.GetCloudAPI(pool.Spec.CloudIntegration); ci != nil {
		return ci.GetAndAssignNewAddress(nodeName, ipName)
	}
	return "", nil
}

func deleteAddressFromCloud(pool *loadbalancing_v1alpha1.EphemeralIPPool, address string) error {
	if pool.Spec.CloudIntegration == nil {
		return nil
	}

	if ci := cloudsIntegration.GetCloudAPI(pool.Spec.CloudIntegration); ci != nil {
		return ci.DeleteAddress(address)
	}
	return nil

}

func allocateAddressOnCloud(pool *loadbalancing_v1alpha1.EphemeralIPPool, address, nodeName string) error {
	if pool.Spec.CloudIntegration == nil {
		return nil
	}

	if ci := cloudsIntegration.GetCloudAPI(pool.Spec.CloudIntegration); ci != nil {
		return ci.AssignIPToServer(address, nodeName)
	}
	return nil
}

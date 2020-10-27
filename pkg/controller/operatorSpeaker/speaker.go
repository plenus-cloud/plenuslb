package operatorspeaker

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"plenus.io/plenuslb/pkg/controller/operator"
	"plenus.io/plenuslb/pkg/controller/utils"
	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
)

// EnsureIPAllocationOnNode adds the address to the right node and make sure it is not on all the others
func EnsureIPAllocationOnNode(nodeName, interfaceName, address string) error {
	operatorsNodes := operator.GetOperatorsList()
	for _, obj := range operatorsNodes {
		podNode, ok := obj.(*v1.Pod)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			continue
		}
		isNodeReady := utils.IsPodReady(podNode)

		node := operator.GetNodeFromOperatorPod(podNode)
		conn, client, err := operator.GrpcClientForNode(node)
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				// Error was not a status error
				klog.Error(err)
			} else {
				klog.Errorf("Error dialoging with the operator %s: code %s message %s", nodeName, st.Code().String(), st.Message())
			}
			return utils.ErrFailedToDialWithOperator
		}
		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if node.NodeName == nodeName {
			if !isNodeReady {
				err := fmt.Errorf("Can't add address to the not ready operator node %s", nodeName)
				klog.Error(err)
				return utils.ErrFailedToDialWithOperator
			}
			_, err = (*client).AddAddress(ctx, &plenuslbV1Alpha1.AddressInfo{Interface: interfaceName, Address: address})
			if err != nil {
				if st, ok := status.FromError(err); ok {
					// Error was a status error
					klog.Errorf("Cannot add address from node. Error dialoging with the operator %s on node %s: code %s message %s", podNode.GetName(), nodeName, st.Code().String(), st.Message())
					return utils.ErrFailedToDialWithOperator
				}
				klog.Error(err)
				return err
			}
			klog.Infof("Added address %s on interface %s of node %s", address, interfaceName, node.NodeName)
		} else if isNodeReady {
			_, err = (*client).RemoveAddress(ctx, &plenuslbV1Alpha1.AddressInfo{Interface: interfaceName, Address: address})
			if err != nil {
				if st, ok := status.FromError(err); ok {
					// Error was a status error
					klog.Errorf("Cannot remove address from node. Error dialoging with the operator %s on node %s: code %s message %s", podNode.GetName(), nodeName, st.Code().String(), st.Message())
				} else {
					klog.Error(err)
				}
			}
			klog.Infof("Removed address %s from interface %s of node %s", address, interfaceName, node.NodeName)
		}
	}

	return nil
}

// RemoveAddressFromNode makes the request to the given perator to remove a specific address
func RemoveAddressFromNode(nodeName, interfaceName, address string) error {
	klog.Infof("Removing address %s from interface %s of node %s", address, interfaceName, nodeName)
	if operatorNode := operator.SearchOperatorByClusterNodeName(nodeName); operatorNode != nil {
		conn, operatorClient, err := operator.GrpcClientForNode(*operatorNode)
		if err != nil {
			klog.Error(err)
			return err
		}
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err = (*operatorClient).RemoveAddress(ctx, &plenuslbV1Alpha1.AddressInfo{
			Address:   address,
			Interface: interfaceName,
		})
		if err != nil {
			klog.Error(err)
			return err
		}
		klog.Infof("Removed address %s from interface %s of node %s", address, interfaceName, nodeName)
	} else {
		// TODO: retry??
		err := fmt.Errorf("Operator for node %s not found", nodeName)
		klog.Error(err)
		return err
	}
	return nil
}

// DoCleanup makes the request at the given oppetator to remove all the not required ips
func DoCleanup(nodeName string, toKeep []*plenuslbV1Alpha1.AddressInfo) error {
	toKeepString := []string{}
	for _, k := range toKeep {
		toKeepString = append(toKeepString, k.GetAddress())
	}
	klog.Infof("Cleaning up addresses on node %s, except %s", nodeName, toKeepString)
	if operatorNode := operator.SearchOperatorByClusterNodeName(nodeName); operatorNode != nil {
		conn, operatorClient, err := operator.GrpcClientForNode(*operatorNode)
		if err != nil {
			klog.Error(err)
			return err
		}
		defer conn.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err = (*operatorClient).Cleanup(ctx, &plenuslbV1Alpha1.CleanupInfo{
			KeepThese: toKeep,
		})
		if err != nil {
			klog.Error(err)
			return err
		}
		klog.Infof("Cleanup on node %s done", nodeName)
	} else {
		err := fmt.Errorf("Operator for node %s not found", nodeName)
		klog.Error(err)
		return err
	}
	return nil
}

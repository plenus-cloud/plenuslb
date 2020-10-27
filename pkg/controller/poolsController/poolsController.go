package poolscontroller

import (
	"reflect"

	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/ephemeralips"
	"plenus.io/plenuslb/pkg/controller/events"
	"plenus.io/plenuslb/pkg/controller/operator"
	"plenus.io/plenuslb/pkg/controller/persistentips"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// Init performs all the startup task for the pools controller
// and registers the callback over the events channels
func Init() {
	events.RegisterOnPersistentPoolDeletedFunc(persistentPoolRemoved)
	events.RegisterOnEphemeralPoolDeletedFunc(ephemeralPoolRemoved)

	events.RegisterOnPersistentPoolModifiedFunc(persistentPoolModified)
	events.RegisterOnEphemeralPoolModifiedFunc(ephemeralPoolModified)
}

func persistentPoolRemoved(pool *loadbalancing_v1alpha1.PersistentIPPool) {

	// check if the removed pool had addAddressesToInterface options
	// -> if yes iterate over pools list
	//    -> if someone still has the option then do nothing
	//    -> if noone has the option than delete the controllers daemonset

	if utils.PersistentPoolHasHostNetworkOption(pool) {
		atListOnePoolHasHostNetworkOption := atLeastOnePoolHasHostNetworkOption()
		if atListOnePoolHasHostNetworkOption && !operator.IsDeployed() {
			operator.DeployOrDie()
		} else if !atListOnePoolHasHostNetworkOption && operator.IsDeployed() {
			operator.Delete()
		}
	}
}

func ephemeralPoolRemoved(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	// check if the removed pool had addAddressesToInterface options
	// -> if yes iterate over pools list
	//    -> if someone still has the option then do nothing
	//    -> if noone has the option than delete the controllers daemonset

	if utils.EphemeralPoolHasHostNetworkOption(pool) {
		atListOnePoolHasHostNetworkOption := atLeastOnePoolHasHostNetworkOption()
		if atListOnePoolHasHostNetworkOption && !operator.IsDeployed() {
			operator.DeployOrDie()
		} else if !atListOnePoolHasHostNetworkOption && operator.IsDeployed() {
			operator.Delete()
		}
	}
}

func ephemeralPoolModified(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	//    -> if someone still has the option then do nothing
	//    -> if noone has the option than delete the controllers daemonset

	atListOnePoolHasHostNetworkOption := atLeastOnePoolHasHostNetworkOption()
	if atListOnePoolHasHostNetworkOption && !operator.IsDeployed() {
		operator.DeployOrDie()
	} else if !atListOnePoolHasHostNetworkOption && operator.IsDeployed() {
		operator.Delete()
	}

}

func persistentPoolModified(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	//    -> if someone still has the option then do nothing
	//    -> if noone has the option than delete the controllers daemonset

	atListOnePoolHasHostNetworkOption := atLeastOnePoolHasHostNetworkOption()
	if atListOnePoolHasHostNetworkOption && !operator.IsDeployed() {
		operator.DeployOrDie()
	} else if !atListOnePoolHasHostNetworkOption && operator.IsDeployed() {
		operator.Delete()
	}
}

// AtLeastOnePoolHasHostNetworkOption checks if at least one pool has the nework option enabled
func atLeastOnePoolHasHostNetworkOption() bool {
	for _, obj := range ephemeralips.GetPoolsList() {
		pool, ok := obj.(*loadbalancing_v1alpha1.EphemeralIPPool)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			continue
		}
		if utils.EphemeralPoolHasHostNetworkOption(pool) {
			klog.Infof("Pool %s has network option", pool.GetName())
			return true
		}
	}
	for _, obj := range persistentips.GetPoolsList() {
		pool, ok := obj.(*loadbalancing_v1alpha1.PersistentIPPool)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			continue
		}
		if utils.PersistentPoolHasHostNetworkOption(pool) {
			klog.Infof("Pool %s has network option", pool.GetName())
			return true
		}
	}
	klog.Infof("No pool has network option")
	return false
}

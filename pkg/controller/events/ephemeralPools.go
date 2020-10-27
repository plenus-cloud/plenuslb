package events

import (
	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

var modifiedEphemeralPoolsChan = make(chan *loadbalancing_v1alpha1.EphemeralIPPool, 100)
var deletedEphemeralPoolsChan = make(chan *loadbalancing_v1alpha1.EphemeralIPPool, 100)

var onEphemeralPoolModifiedCB []*func(pool *loadbalancing_v1alpha1.EphemeralIPPool)
var onEphemeralPoolDeletedCB []*func(pool *loadbalancing_v1alpha1.EphemeralIPPool)

// EphemeralPoolModified pushes the modified ephemeral pool into the modified ephemeral pool event channel
func EphemeralPoolModified(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	modifiedEphemeralPoolsChan <- pool.DeepCopy()
}

// EphemeralPoolDeleted pushes the deleted ephemeral pool into the deleted ephemeral pool event channel
func EphemeralPoolDeleted(pool *loadbalancing_v1alpha1.EphemeralIPPool) {
	deletedEphemeralPoolsChan <- pool.DeepCopy()
}

// RegisterOnEphemeralPoolModifiedFunc registers a callback to be fired when an ephemeral pool is modified
func RegisterOnEphemeralPoolModifiedFunc(cb func(pool *loadbalancing_v1alpha1.EphemeralIPPool)) {
	onEphemeralPoolModifiedCB = append(onEphemeralPoolModifiedCB, &cb)
}

// RegisterOnEphemeralPoolDeletedFunc registers a callback to be fired when an ephemeral pool is deleted
func RegisterOnEphemeralPoolDeletedFunc(cb func(pool *loadbalancing_v1alpha1.EphemeralIPPool)) {
	onEphemeralPoolDeletedCB = append(onEphemeralPoolDeletedCB, &cb)
}

// ListenModifiedEphemeralPoolsChan strarts a listener on the channel of modified ephemeral pools
// and fires, ad each event, all the callbacks previously registered for this kind of event
func ListenModifiedEphemeralPoolsChan(stopCh chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				klog.Warning("Stopping ListenModifiedEphemeralPoolsChan goroutine")
				return
			case modified := <-modifiedEphemeralPoolsChan:
				if len(onEphemeralPoolModifiedCB) > 0 {
					for _, cb := range onEphemeralPoolModifiedCB {
						(*cb)(modified)
					}
				} else {
					klog.Error("onEphemeralPoolModifiedCB is nil, cannot process persistent pool modified event")
				}
			}
		}
	}()
}

// ListenDeletedEphemeralPoolsChan strarts a listener on the channel of deleted ephemeral pools
// and fires, ad each event, all the callbacks previously registered for this kind of event
func ListenDeletedEphemeralPoolsChan(stopCh chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				klog.Warning("Stopping ListenDeletedEphemeralPoolsChan goroutine")
				return
			case deleted := <-deletedEphemeralPoolsChan:
				if len(onEphemeralPoolDeletedCB) > 0 {
					for _, cb := range onEphemeralPoolDeletedCB {
						(*cb)(deleted)
					}
				} else {
					klog.Error("onEphemeralPoolDeletedCB is nil, cannot process persistent pool deleted event")
				}
			}
		}
	}()
}

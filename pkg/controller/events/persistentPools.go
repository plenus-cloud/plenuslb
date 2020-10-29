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

package events

import (
	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

var modifiedPoolsChan = make(chan *loadbalancing_v1alpha1.PersistentIPPool, 100)
var deletedPoolsChan = make(chan *loadbalancing_v1alpha1.PersistentIPPool, 100)

var onPoolModifiedCB []*func(pool *loadbalancing_v1alpha1.PersistentIPPool)
var onPoolDeletedCB []*func(pool *loadbalancing_v1alpha1.PersistentIPPool)

// PersistentPoolModified pushes the modified persistent pool into the modified persistent pool event channel
func PersistentPoolModified(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	modifiedPoolsChan <- pool.DeepCopy()
}

// PersistentPoolDeleted pushes the deleted persistent pool into the deleted persistent pool event channel
func PersistentPoolDeleted(pool *loadbalancing_v1alpha1.PersistentIPPool) {
	deletedPoolsChan <- pool.DeepCopy()
}

// RegisterOnPersistentPoolModifiedFunc registers a callback to be fired when an persistent pool is modified
func RegisterOnPersistentPoolModifiedFunc(cb func(pool *loadbalancing_v1alpha1.PersistentIPPool)) {
	onPoolModifiedCB = append(onPoolModifiedCB, &cb)
}

// RegisterOnPersistentPoolDeletedFunc registers a callback to be fired when an persistent pool is deleted
func RegisterOnPersistentPoolDeletedFunc(cb func(pool *loadbalancing_v1alpha1.PersistentIPPool)) {
	onPoolDeletedCB = append(onPoolDeletedCB, &cb)
}

// ListenModifiedPersistentPoolsChan strarts a listener on the channel of modified persistent pools
// and fires, ad each event, all the callbacks previously registered for this kind of event
func ListenModifiedPersistentPoolsChan(stopCh chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				klog.Warning("Stopping ListenModifiedPersistentPoolsChan goroutine")
				return
			case modified := <-modifiedPoolsChan:
				if len(onPoolModifiedCB) > 0 {
					for _, cb := range onPoolModifiedCB {
						(*cb)(modified)
					}
				} else {
					klog.Error("onPoolModifiedCB is nil, cannot process persistent pool modified event")
				}
			}
		}
	}()
}

// ListenDeletedPersistentPoolsChan strarts a listener on the channel of deleted persistent pools
// and fires, ad each event, all the callbacks previously registered for this kind of event
func ListenDeletedPersistentPoolsChan(stopCh chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				klog.Warning("Stopping ListenDeletedPersistentPoolsChan goroutine")
				return
			case deleted := <-deletedPoolsChan:
				if len(onPoolDeletedCB) > 0 {
					for _, cb := range onPoolDeletedCB {
						(*cb)(deleted)
					}
				} else {
					klog.Error("onPoolDeletedCB is nil, cannot process persistent pool deleted event")
				}
			}
		}
	}()
}

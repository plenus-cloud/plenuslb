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

package observer

import (
	"context"
	"sync"
	"time"

	"github.com/vishvananda/netlink"

	"k8s.io/klog"

	"plenus.io/plenuslb/pkg/operator/network"
	"plenus.io/plenuslb/pkg/operator/utils"
	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
)

// this is the list of all ip assigned to the operator
var assignedAddressesList = []*plenuslbV1Alpha1.AddressInfo{}

// mutex for address manipulation
var addressesLock = &sync.Mutex{}

// do observer business
// subscribe to addresses update and watch them
func Run(doneAddrUpdate chan struct{}) {
	// subscribe to ip addresses update
	chAddrUpdate := make(chan netlink.AddrUpdate)
	err := subscribeAddrUpdate(chAddrUpdate, doneAddrUpdate)
	if err != nil {
		klog.Fatalf("Fatal error received during AddrSubscribe subscription: %v", err)
	}

	// watch for addresses update
	go watchAddrUpdate(chAddrUpdate)
}

// AddAddress adds given address to specific interface
func AddAddress(info *plenuslbV1Alpha1.AddressInfo) error {
	addressesLock.Lock()
	defer addressesLock.Unlock()
	err := network.AddAddress(info.GetInterface(), info.GetAddress())
	if err != nil {
		return err
	}
	if !utils.ContainsAddressInfo(assignedAddressesList, info.GetInterface(), info.GetAddress()) {
		// add address to assignedAddressesList
		assignedAddressesList = append(assignedAddressesList, info)
	}
	return nil
}

// RemoveAddress removes given address from specific interface
func RemoveAddress(info *plenuslbV1Alpha1.AddressInfo) error {
	addressesLock.Lock()
	defer addressesLock.Unlock()
	err := network.DeleteAddress(info.GetInterface(), info.GetAddress())
	if err != nil {
		return err
	}

	// remove address from the list of managed addresses
	newAddressList := []*plenuslbV1Alpha1.AddressInfo{}
	for _, currentAddress := range assignedAddressesList {
		if currentAddress.GetAddress() != info.GetAddress() || currentAddress.GetInterface() != info.GetInterface() {
			newAddressList = append(newAddressList, currentAddress)
		}
	}
	assignedAddressesList = newAddressList

	return nil
}

// Cleanup removes all addresses managed by plenuslb from all interfaces
func Cleanup(keepThese []*plenuslbV1Alpha1.AddressInfo) error {
	addressesLock.Lock()
	defer addressesLock.Unlock()

	// apply network.Cleanup
	if err := network.Cleanup(keepThese); err != nil {
		return err
	}

	// restore missing addresses
	for _, info := range keepThese {
		err := network.AddAddress(info.GetInterface(), info.GetAddress())
		if err != nil {
			return err
		}
	}

	// if succeed update assignedAddressesList
	assignedAddressesList = keepThese
	return nil
}

// watch for ip addresses update messages
// every time a message arrives if it regards an address that the operator should have assigned to the interface
// the operator will check if the address is still assigned
// if it is not it will restore it.
// this avoid losing the ip due to some external event
// ex. the system updating package systemd on ubuntu distro
func watchAddrUpdate(chAddrUpdate chan netlink.AddrUpdate) {
	for range chAddrUpdate {
		// process the address update
		// wihich address has been updated is not important
		// we will search for "lost" addresses whatever
		processAddressUpdate()
	}
}

// process the address update
// cycle on assignedAddressesList and
// search for addresses that are not assigned to the required interface
func processAddressUpdate() {
	addressesLock.Lock()
	defer addressesLock.Unlock()

	// the lock inside this function could block the functions called by grpc
	// if we cannot terminate within 5 seconds we will terminate the process
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chanProcessAddressUpdate := make(chan error, 1)
	go func() {
		for _, currentAddress := range assignedAddressesList {
			addressFound, err := network.IsAddressOnInterface(currentAddress)
			if err != nil {
				klog.Error(err)
				continue
			}

			// if address wasn't found add it to interface
			if !addressFound {
				klog.Infof("Found missing address %s on interface %s", currentAddress.GetAddress(), currentAddress.GetInterface())
				err = network.AddAddress(currentAddress.GetInterface(), currentAddress.GetAddress())
				if err != nil {
					klog.Errorf("Failed to restore missing address %s on interface %s", currentAddress.GetAddress(), currentAddress.GetInterface())
					continue
				}
			}
		}

		chanProcessAddressUpdate <- nil
	}()

	select {
	case <-ctx.Done():
		klog.Infof("Timeout in processAddressUpdate")
		// log a fatal error and terminate
		// we were not able to process the message and restore correct addresses within deadline
		// by killing the operator we notify the controller that things went wrong in a bad way
		klog.Fatal(ctx.Err())
	case <-chanProcessAddressUpdate:
	}
}

// subscribe to ip addresses update
func subscribeAddrUpdate(chAddrUpdate chan netlink.AddrUpdate, doneAddrUpdate chan struct{}) error {
	return netlink.AddrSubscribe(chAddrUpdate, doneAddrUpdate)
}

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

package network

import (
	"net"
	"syscall"

	"github.com/vishvananda/netlink"

	"k8s.io/klog"

	"plenus.io/plenuslb/pkg/operator/utils"
	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
)

const labelName = "pllb"

// Cleanup cleans all the configurations done except those provided
func Cleanup(keepThese []*plenuslbV1Alpha1.AddressInfo) error {
	// cleanup ipvs rules by flush
	klog.Infof("Cleaning up addresses permanently")
	links, err := netlink.LinkList()
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, link := range links {
		address, err := netlink.AddrList(link, 0)
		if err != nil {
			klog.Error(err)
			return err
		}
		for _, addres := range address {
			if addres.Label == composeLinkLabel(link) && !utils.ContainsAddressInfo(keepThese, link.Attrs().Name, addres.IP.String()) {
				klog.Infof("Address %s will be deleted from interface %s", addres.IP.String(), link.Attrs().Name)

				err := netlink.AddrDel(link, &addres)
				if err != nil {
					klog.Error(err)
				}
			}
		}
	}
	klog.Infof("Successfully cleaned addresses added by plenuslb")
	return nil
}

func composeLinkLabel(link netlink.Link) string {
	return link.Attrs().Name + ":" + labelName
}

// DeleteAddress removes an adress from a network interface, if exists
func DeleteAddress(netInterfaceName, address string) error {
	netInterface, err := getInterfaceByName(netInterfaceName)
	if err != nil {
		klog.Errorf("Failed to get network interface %s due to %s", netInterfaceName, err.Error())
		return err
	}

	addresses, err := netlink.AddrList(netInterface, 0)
	if err != nil {
		klog.Error(err)
		return err
	}
	for _, addres := range addresses {
		if addres.IP.String() == address {
			klog.Infof("Address %s will be deleted from interface %s", addres.IP.String(), netInterface.Attrs().Name)

			err := netlink.AddrDel(netInterface, &addres)
			if err != nil {
				klog.Error(err)
				return err
			}
			return nil
		}
	}
	klog.Infof("Deleted address %s from interface %s", address, netInterfaceName)
	return nil
}

// AddAddress adds an adress to a network interface
func AddAddress(netInterfaceName, address string) error {
	netInterface, err := getInterfaceByName(netInterfaceName)
	if err != nil {
		klog.Errorf("Failed to get network interface %s due to %s", netInterfaceName, err.Error())
		return err
	}

	eip := &netlink.Addr{IPNet: &net.IPNet{
		IP:   net.ParseIP(address),
		Mask: net.IPv4Mask(255, 255, 255, 255)},
		Scope: syscall.RT_SCOPE_LINK,
		Label: composeLinkLabel(netInterface),
	}
	err = netlink.AddrAdd(netInterface, eip)
	if err != nil && err.Error() != "file exists" {
		klog.Errorf("Failed to assign external ip %s to interface %s due to %s", address, netInterfaceName, err.Error())
		return err
	}
	klog.Infof("Added address %s to interface %s", address, netInterfaceName)
	return nil
}

func getInterfaceByName(netInterfaceName string) (netlink.Link, error) {
	netInterface, err := netlink.LinkByName(netInterfaceName)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	klog.Infof("Successfully got interface %s", netInterfaceName)
	return netInterface, nil
}

func IsAddressOnInterface(addressToBeChecked *plenuslbV1Alpha1.AddressInfo) (bool, error) {
	// verify if address is assigned to interface
	// get interface
	link, err := getInterfaceByName(addressToBeChecked.GetInterface())
	if err != nil {
		return false, err
	}
	// get all addresses on interface
	realAddressList, err := netlink.AddrList(link, 0)
	if err != nil {
		return false, err
	}
	// compare addresses on interface with currentAddress
	// there is no need to compare interface name as we have already got only the addresses on that interface
	for _, address := range realAddressList {
		if (addressToBeChecked.GetAddress() == address.IP.String()) {
			return true, nil
		}
	}
	return false, nil
}
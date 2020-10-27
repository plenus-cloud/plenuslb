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

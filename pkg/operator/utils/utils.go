package utils

import (
	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
)

// ContainsAddressInfo tells whether a contains x.
func ContainsAddressInfo(a []*plenuslbV1Alpha1.AddressInfo, interfaceName, address string) bool {
	for _, n := range a {
		if address == n.GetAddress() && n.GetInterface() == interfaceName {
			return true
		}
	}
	return false
}

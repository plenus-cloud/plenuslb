package utils

import loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"

// ContainsAddress tells whether a contains x.
func ContainsAddress(a []*loadbalancing_v1alpha1.IPAllocationAddresses, x string) (bool, *loadbalancing_v1alpha1.IPAllocationAddresses) {
	for _, n := range a {
		if x == n.Address {
			return true, n
		}
	}
	return false, nil
}

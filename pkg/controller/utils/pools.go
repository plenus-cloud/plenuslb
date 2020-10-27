package utils

import (
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

// PersistentPoolHasHostNetworkOption checks if the given pool ha the nework option enabled
func PersistentPoolHasHostNetworkOption(pool *loadbalancing_v1alpha1.PersistentIPPool) bool {
	if pool.Spec.Options != nil && pool.Spec.Options.HostNetworkInterface != nil && pool.Spec.Options.HostNetworkInterface.AddAddressesToInterface {
		return true
	}
	return false
}

// EphemeralPoolHasHostNetworkOption checks if the given pool ha the nework option enabled
func EphemeralPoolHasHostNetworkOption(pool *loadbalancing_v1alpha1.EphemeralIPPool) bool {
	if pool.Spec.Options != nil && pool.Spec.Options.HostNetworkInterface != nil && pool.Spec.Options.HostNetworkInterface.AddAddressesToInterface {
		return true
	}
	return false
}

// PersistentPoolHasCloudIntegrationOption check if the give pool has the integration with a cloud
func PersistentPoolHasCloudIntegrationOption(pool *loadbalancing_v1alpha1.PersistentIPPool) (bool, string) {
	if pool.Spec.CloudIntegration == nil {
		return false, ""
	}

	if pool.Spec.CloudIntegration.Hetzner != nil {
		return true, "hetzner"
	}
	return false, ""
}

// EphemeralPoolHasCloudIntegrationOption check if the give pool has the integration with a cloud
func EphemeralPoolHasCloudIntegrationOption(pool *loadbalancing_v1alpha1.EphemeralIPPool) (bool, string) {
	if pool.Spec.CloudIntegration == nil {
		return false, ""
	}

	if pool.Spec.CloudIntegration.Hetzner != nil {
		return true, "hetzner"
	}
	return false, ""
}

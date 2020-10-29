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

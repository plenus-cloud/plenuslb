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

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

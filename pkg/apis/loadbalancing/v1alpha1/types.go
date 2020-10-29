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

package v1alpha1

// PoolOptions is the type for IPPoolSpec options field
type PoolOptions struct {
	HostNetworkInterface *HostNetworkInterfaceOptions `json:"hostNetworkInterface"`
}

// HostNetworkInterfaceOptions are the options required if you wish to add the ips to the interface
type HostNetworkInterfaceOptions struct {
	AddAddressesToInterface bool   `json:"addAddressesToInterface"`
	InterfaceName           string `json:"interfaceName"`
}

// CloudIntegrations is the type for IPPoolSpec cloudIntegration field
type CloudIntegrations struct {
	Hetzner *HetznerCloud `json:"hetzner"`
}

// HetznerCloud is the type for CloudIntegrations hetzner provider
type HetznerCloud struct {
	Token string `json:"token"`
}

// IPPoolStatus is the IPPool status
type IPPoolStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

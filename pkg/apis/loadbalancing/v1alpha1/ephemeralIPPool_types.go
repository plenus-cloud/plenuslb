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

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EphemeralIPPool is the top level type
type EphemeralIPPool struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               EphemeralIPPoolSpec `json:"spec"`
	Status             IPPoolStatus        `json:"status,omitempty"`
}

// EphemeralIPPoolSpec is the spec type for EphemeralIPPool
type EphemeralIPPoolSpec struct {
	AllowedNamespaces []string           `json:"allowedNamespaces"`
	CloudIntegration  *CloudIntegrations `json:"cloudIntegration,omitempty"`
	Options           *PoolOptions       `json:"options,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EphemeralIPPoolList defines the list of ippools
type EphemeralIPPoolList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []*EphemeralIPPool `json:"items"`
}

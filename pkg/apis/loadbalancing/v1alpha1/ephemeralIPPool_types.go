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

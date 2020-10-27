package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PersistentIPPool is the top level type
type PersistentIPPool struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               PersistentIPPoolSpec `json:"spec"`
	Status             IPPoolStatus         `json:"status,omitempty"`
}

// PersistentIPPoolSpec is the spec type for PersistentIPPool
type PersistentIPPoolSpec struct {
	Addresses         []string           `json:"addresses"`
	AllowedNamespaces []string           `json:"allowedNamespaces"`
	CloudIntegration  *CloudIntegrations `json:"cloudIntegration,omitempty"`
	Options           *PoolOptions       `json:"options,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PersistentIPPoolList defines the list of ippools
type PersistentIPPoolList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []*PersistentIPPool `json:"items"`
}

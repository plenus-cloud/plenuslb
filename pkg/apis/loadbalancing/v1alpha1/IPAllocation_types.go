package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAllocation is the top level type
type IPAllocation struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               IPAllocationSpec   `json:"spec"`
	Status             IPAllocationStatus `json:"status,omitempty"`
}

// IPType is the type of the ip
type IPType string

const (
	// EphemeralIP are ips that are dinamically added and removed pìfrom cloud
	EphemeralIP IPType = "ephemeral"
	// PersistentIP are ips that are just assigned or unassigned to a machine
	PersistentIP IPType = "persistent"
)

// IPAllocationSpec are the specs of ip allocation
type IPAllocationSpec struct {
	Type        IPType                   `json:"type"`
	Allocations []*IPAllocationAddresses `json:"allocations"`
}

// IPAllocationAddresses are the allocated address details
type IPAllocationAddresses struct {
	Address          string `json:"address"`
	NetworkInterface string `json:"networkInterface,omitempty"`
	NodeName         string `json:"nodeName,omitempty"`
	CloudProvider    string `json:"cloudProvider,omitempty"`
	Pool             string `json:"pool,omitempty"`
}

// AllocationStatus are the status of allocations
type AllocationStatus string

const (
	// AllocationStatusSuccess is the status whe the allocation is done
	AllocationStatusSuccess AllocationStatus = "success"
	// AllocationStatusError is the status when the allocation is error
	// the allocation with this state will be retried for 2 days with a capped backoff
	AllocationStatusError AllocationStatus = "error"
	// AllocationStatusFailed is the status when the allocation is failed
	// the allocation with this state won't be retried
	AllocationStatusFailed AllocationStatus = "failed"
	// AllocationStatusNodeError is the stattus when an allocation is on a node_error node
	// the allocation with this state will be relocated
	AllocationStatusNodeError AllocationStatus = "node_error"
	// AllocationStatusPending is the status when the allocation is waiting to be processed
	// the allocation with this state will be allocated
	AllocationStatusPending AllocationStatus = "pending"
	// AllocationStatusAddrDeleted is the status used when at least one addr has been removed from poll
	AllocationStatusAddrDeleted AllocationStatus = "address_deleted_from_pool"
)

// IPAllocationStatus is the IPAllocation status
type IPAllocationStatus struct {
	State   AllocationStatus `json:"state,omitempty"`
	Message string           `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAllocationList defines the list of ippools
type IPAllocationList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []IPAllocation `json:"items"`
}

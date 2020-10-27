package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// IPAllocationCRDPlural is the plural name of the IPAllocationCRD
	IPAllocationCRDPlural string = "ipallocations"
	// FullIPAllocationCRDName is the full name of the IPAllocationCRD
	FullIPAllocationCRDName string = IPAllocationCRDPlural + "." + CRDGroup
)

func addIPAllocationKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&IPAllocation{},
		&IPAllocationList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

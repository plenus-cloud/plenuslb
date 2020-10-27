package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// EphemeralIPPoolCRDPlural is the plural name of the EphemeralIPPoolCRD
	EphemeralIPPoolCRDPlural string = "ephemeralippools"
	// FullEphemeralIPPoolCRDName is the full name of the EphemeralIPPoolCRD
	FullEphemeralIPPoolCRDName string = EphemeralIPPoolCRDPlural + "." + CRDGroup
)

func addEphemeralIPPoolKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&EphemeralIPPool{},
		&EphemeralIPPoolList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

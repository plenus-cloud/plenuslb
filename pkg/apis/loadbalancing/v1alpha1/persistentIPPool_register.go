package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// PersistentIPPoolCRDPlural is the plural name of the PersistentIPPoolCRD
	PersistentIPPoolCRDPlural string = "persistentippools"
	// FullPersistentIPPoolCRDName is the full name of the PersistentIPPoolCRD
	FullPersistentIPPoolCRDName string = PersistentIPPoolCRDPlural + "." + CRDGroup
)

func addPersistentIPPoolKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&PersistentIPPool{},
		&PersistentIPPoolList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

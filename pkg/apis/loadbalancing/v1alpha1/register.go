package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// CRDGroup is the group of the CRD
	CRDGroup string = "loadbalancing.plenus.io"
	// CRDVersion is the version of the CRD
	CRDVersion string = "v1alpha1"
)

var (
	// SchemeBuilder is the builder of the CRD
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	// AddToScheme is the cheme adder of the CRD
	AddToScheme = localSchemeBuilder.AddToScheme
)

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// SchemeGroupVersion is the scheme version of the CRD
var SchemeGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

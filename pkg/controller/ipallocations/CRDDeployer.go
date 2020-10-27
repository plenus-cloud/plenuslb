package ipallocations

import (
	"reflect"
	"strings"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	crddeployer "plenus.io/plenuslb/pkg/controller/CRDDeployer"
)

// CreateOrUpdateCRD create or opdate the IPAllocation CRD
func CreateOrUpdateCRD() error {
	return crddeployer.CreateOrUpdateCRD(loadbalancing_v1alpha1.FullIPAllocationCRDName, getV1Definition())
}

func getV1Definition() *apiextv1.CustomResourceDefinition {
	return &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: loadbalancing_v1alpha1.FullIPAllocationCRDName,
		},
		Spec: apiextv1.CustomResourceDefinitionSpec{
			Versions: []apiextv1.CustomResourceDefinitionVersion{
				{
					Name:    loadbalancing_v1alpha1.CRDVersion,
					Served:  true,
					Storage: true,
					Schema:  loadbalancing_v1alpha1.GetIPAllocationValidationSchemaV1(),
					Subresources: &apiextv1.CustomResourceSubresources{
						Status: &apiextv1.CustomResourceSubresourceStatus{},
					},
					AdditionalPrinterColumns: []apiextv1.CustomResourceColumnDefinition{
						{
							Name:     "Type",
							Type:     "string",
							JSONPath: ".spec.type",
						},
						{
							Name:     "State",
							Type:     "string",
							JSONPath: ".status.state",
						},
						{
							Name:     "Message",
							Type:     "string",
							JSONPath: ".status.message",
						},
						{
							Name:     "Age",
							Type:     "date",
							JSONPath: ".metadata.creationTimestamp",
						},
					},
				},
			},
			Group: loadbalancing_v1alpha1.CRDGroup,
			Scope: apiextv1.NamespaceScoped,
			Names: apiextv1.CustomResourceDefinitionNames{
				Singular:   strings.ToLower(reflect.TypeOf(loadbalancing_v1alpha1.IPAllocation{}).Name()),
				Plural:     loadbalancing_v1alpha1.IPAllocationCRDPlural,
				Kind:       reflect.TypeOf(loadbalancing_v1alpha1.IPAllocation{}).Name(),
				Categories: []string{"all"},
				ShortNames: []string{"ipa"},
			},
			PreserveUnknownFields: false,
		},
	}
}

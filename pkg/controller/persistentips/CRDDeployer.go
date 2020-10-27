package persistentips

import (
	"reflect"
	"strings"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	crddeployer "plenus.io/plenuslb/pkg/controller/CRDDeployer"
)

// CreateOrUpdateCRD create or opdate the PersistentIPPools CRD
func CreateOrUpdateCRD() error {
	return crddeployer.CreateOrUpdateCRD(loadbalancing_v1alpha1.FullPersistentIPPoolCRDName, getV1Definition())
}

func getV1Definition() *apiextv1.CustomResourceDefinition {
	return &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: loadbalancing_v1alpha1.FullPersistentIPPoolCRDName,
		},
		Spec: apiextv1.CustomResourceDefinitionSpec{
			Versions: []apiextv1.CustomResourceDefinitionVersion{
				{
					Name:    loadbalancing_v1alpha1.CRDVersion,
					Served:  true,
					Storage: true,
					Schema:  loadbalancing_v1alpha1.GetPersistentIPPoolValidationSchemaV1(),
				},
			},
			Group: loadbalancing_v1alpha1.CRDGroup,
			Scope: apiextv1.ClusterScoped,
			Names: apiextv1.CustomResourceDefinitionNames{
				Singular: strings.ToLower(reflect.TypeOf(loadbalancing_v1alpha1.PersistentIPPool{}).Name()),
				Plural:   loadbalancing_v1alpha1.PersistentIPPoolCRDPlural,
				Kind:     reflect.TypeOf(loadbalancing_v1alpha1.PersistentIPPool{}).Name(),
			},
			PreserveUnknownFields: false,
		},
	}
}

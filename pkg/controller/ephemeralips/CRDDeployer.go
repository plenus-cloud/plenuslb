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

package ephemeralips

import (
	"reflect"
	"strings"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	crddeployer "plenus.io/plenuslb/pkg/controller/CRDDeployer"
)

// CreateOrUpdateCRD create or opdate the EphemeralIPPools CRD
func CreateOrUpdateCRD() error {
	return crddeployer.CreateOrUpdateCRD(loadbalancing_v1alpha1.FullEphemeralIPPoolCRDName, getV1Definition())
}

// CreateCRD creates the ippools CRD
func getV1Definition() *apiextv1.CustomResourceDefinition {
	return &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: loadbalancing_v1alpha1.FullEphemeralIPPoolCRDName,
		},
		Spec: apiextv1.CustomResourceDefinitionSpec{
			Group: loadbalancing_v1alpha1.CRDGroup,
			Versions: []apiextv1.CustomResourceDefinitionVersion{
				{
					Name:    loadbalancing_v1alpha1.CRDVersion,
					Served:  true,
					Storage: true,
					Schema:  loadbalancing_v1alpha1.GetEphemeralIPPoolValidationSchemaV1(),
				},
			},
			Scope: apiextv1.ClusterScoped,
			Names: apiextv1.CustomResourceDefinitionNames{
				Singular: strings.ToLower(reflect.TypeOf(loadbalancing_v1alpha1.EphemeralIPPool{}).Name()),
				Plural:   loadbalancing_v1alpha1.EphemeralIPPoolCRDPlural,
				Kind:     reflect.TypeOf(loadbalancing_v1alpha1.EphemeralIPPool{}).Name(),
			},
			PreserveUnknownFields: false,
		},
	}
}

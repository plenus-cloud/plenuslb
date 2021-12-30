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

package v1alpha1

import (
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// GetPersistentIPPoolValidationSchemaV1 returns the validation schema con PersistentIPPool CRD
func GetPersistentIPPoolValidationSchemaV1() *apiextv1.CustomResourceValidation {
	minArrayLength := int64(1)
	return &apiextv1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextv1.JSONSchemaProps{
			Required: []string{"spec"},
			Type:     "object",
			Properties: map[string]apiextv1.JSONSchemaProps{
				"spec": {
					Type:     "object",
					Required: []string{"addresses"},
					Properties: map[string]apiextv1.JSONSchemaProps{
						"addresses": {
							AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
								Allows: false,
							},
							Type: "array",
							Items: &apiextv1.JSONSchemaPropsOrArray{
								Schema: &apiextv1.JSONSchemaProps{
									AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
										Allows: false,
									},
									Type: "string",
								},
							},
							MinLength: &minArrayLength,
						},
						"allowedNamespaces": {
							AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
								Allows: false,
							},
							Type: "array",
							Items: &apiextv1.JSONSchemaPropsOrArray{
								Schema: &apiextv1.JSONSchemaProps{
									AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
										Allows: false,
									},
									Type: "string",
								},
							},
							MinLength: &minArrayLength,
						},
						"cloudIntegration": {
							Type: "object",
							Properties: map[string]apiextv1.JSONSchemaProps{
								"hetzner": {
									Type:     "object",
									Required: []string{"token"},
									Properties: map[string]apiextv1.JSONSchemaProps{
										"token": {
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "string",
										},
									},
								},
							},
							OneOf: []apiextv1.JSONSchemaProps{
								{
									Required: []string{"hetzner"},
								},
							},
						},
						"options": {
							Type: "object",
							Properties: map[string]apiextv1.JSONSchemaProps{
								"hostNetworkInterface": {
									Type:     "object",
									Required: []string{"addAddressesToInterface", "interfaceName"},
									Properties: map[string]apiextv1.JSONSchemaProps{
										"addAddressesToInterface": {
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "boolean",
										},
										"interfaceName": {
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "string",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

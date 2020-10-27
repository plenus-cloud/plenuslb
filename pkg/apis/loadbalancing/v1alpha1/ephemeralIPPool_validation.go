package v1alpha1

import (
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// GetEphemeralIPPoolValidationSchemaV1 returns the validation schema con EphemeralIPPool CRD
func GetEphemeralIPPoolValidationSchemaV1() *apiextv1.CustomResourceValidation {
	var minArrayLength int64
	minArrayLength = 1
	return &apiextv1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextv1.JSONSchemaProps{
			Required: []string{"spec"},
			Type:     "object",
			Properties: map[string]apiextv1.JSONSchemaProps{
				"spec": apiextv1.JSONSchemaProps{
					Type:     "object",
					Required: []string{"cloudIntegration"},
					Properties: map[string]apiextv1.JSONSchemaProps{
						"allowedNamespaces": apiextv1.JSONSchemaProps{
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
						"cloudIntegration": apiextv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextv1.JSONSchemaProps{
								"hetzner": apiextv1.JSONSchemaProps{
									Type:     "object",
									Required: []string{"token"},
									Properties: map[string]apiextv1.JSONSchemaProps{
										"token": apiextv1.JSONSchemaProps{
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "string",
										},
									},
								},
							},
							OneOf: []apiextv1.JSONSchemaProps{
								apiextv1.JSONSchemaProps{
									Required: []string{"hetzner"},
								},
							},
						},
						"options": apiextv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextv1.JSONSchemaProps{
								"hostNetworkInterface": apiextv1.JSONSchemaProps{
									Type:     "object",
									Required: []string{"addAddressesToInterface", "interfaceName"},
									Properties: map[string]apiextv1.JSONSchemaProps{
										"addAddressesToInterface": apiextv1.JSONSchemaProps{
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "boolean",
										},
										"interfaceName": apiextv1.JSONSchemaProps{
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

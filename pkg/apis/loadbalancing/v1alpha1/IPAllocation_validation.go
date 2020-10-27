package v1alpha1

import (
	"fmt"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// GetIPAllocationValidationSchemaV1 returns the validation schema con IPAllocation CRD
func GetIPAllocationValidationSchemaV1() *apiextv1.CustomResourceValidation {
	return &apiextv1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextv1.JSONSchemaProps{
			Required: []string{"spec"},
			Type:     "object",
			Properties: map[string]apiextv1.JSONSchemaProps{
				"spec": apiextv1.JSONSchemaProps{
					Required: []string{"allocations", "type"},
					Type:     "object",
					Properties: map[string]apiextv1.JSONSchemaProps{
						"type": apiextv1.JSONSchemaProps{
							AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
								Allows: false,
							},
							Type: "string",
							Enum: []apiextv1.JSON{
								{
									Raw: []byte(fmt.Sprintf(`"%s"`, EphemeralIP)),
								},
								{
									Raw: []byte(fmt.Sprintf(`"%s"`, PersistentIP)),
								},
							},
						},
						"allocations": {
							Type: "array",
							Items: &apiextv1.JSONSchemaPropsOrArray{
								Schema: &apiextv1.JSONSchemaProps{
									Type:     "object",
									Required: []string{"address", "pool"},
									Properties: map[string]apiextv1.JSONSchemaProps{
										"address": apiextv1.JSONSchemaProps{
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "string",
										},
										"pool": apiextv1.JSONSchemaProps{
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "string",
										},
										"networkInterface": apiextv1.JSONSchemaProps{
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "string",
										},
										"nodeName": apiextv1.JSONSchemaProps{
											AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
												Allows: false,
											},
											Type: "string",
										},
										"cloudProvider": apiextv1.JSONSchemaProps{
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
				"status": apiextv1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextv1.JSONSchemaProps{
						"state": apiextv1.JSONSchemaProps{
							AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
								Allows: false,
							},
							Type: "string",
							Enum: []apiextv1.JSON{
								{
									Raw: []byte(fmt.Sprintf(`"%s"`, AllocationStatusSuccess)),
								},
								{
									Raw: []byte(fmt.Sprintf(`"%s"`, AllocationStatusError)),
								},
								{
									Raw: []byte(fmt.Sprintf(`"%s"`, AllocationStatusNodeError)),
								},
								{
									Raw: []byte(fmt.Sprintf(`"%s"`, AllocationStatusPending)),
								},
								{
									Raw: []byte(fmt.Sprintf(`"%s"`, AllocationStatusAddrDeleted)),
								},
								{
									Raw: []byte(fmt.Sprintf(`"%s"`, AllocationStatusFailed)),
								},
							},
						},
						"message": apiextv1.JSONSchemaProps{
							AdditionalProperties: &apiextv1.JSONSchemaPropsOrBool{
								Allows: false,
							},
							Type: "string",
						},
					},
				},
			},
		},
	}
}

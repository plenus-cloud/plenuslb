package clouds

import (
	"k8s.io/klog"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/clouds/hetzner"
)

// CloudAPI is the interface of each cloud integration
type CloudAPI interface {
	AssignIPToServer(address, serverName string) error
	UnassignIP(address string) error
	GetAndAssignNewAddress(serverName, ipName string) (string, error)
	DeleteAddress(address string) error
}

// Clouds is the interface of the clouds utilities
type Clouds interface {
	GetCloudAPI(cloudIntegrationOpts *loadbalancing_v1alpha1.CloudIntegrations) CloudAPI
}

// Integration contains the declarations of all the utilities for the integrations with the cloud
type Integration struct{}

// GetCloudAPI returns the right cloud api instance according to what is declared in the pool
func (c *Integration) GetCloudAPI(cloudIntegrationOpts *loadbalancing_v1alpha1.CloudIntegrations) CloudAPI {
	if cloudIntegrationOpts.Hetzner != nil {
		return &hetzner.API{
			Token: cloudIntegrationOpts.Hetzner.Token,
		}
	}

	klog.Errorf("Failed to get cloud API for %v", *cloudIntegrationOpts)
	return nil
}

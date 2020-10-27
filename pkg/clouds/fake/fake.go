package fake

import (
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/clouds"
)

// cloudAPI is a silly implementation of a cloud api for testing purposes
type cloudAPI struct{}

// AssignIPToServer is a silly implementation of the function that assigns an existing ip to a server on the cloud
func (c *cloudAPI) AssignIPToServer(address, serverName string) error {
	return nil
}

// UnassignIP is a silly implementation of the function that unassigns an ip from a server on the cloud
func (c *cloudAPI) UnassignIP(address string) error {
	return nil
}

// GetAndAssignNewAddress is a silly implementation of the function that obtains and assigns an ip to a server on the cloud
func (c *cloudAPI) GetAndAssignNewAddress(serverName, ipName string) (string, error) {
	return "1.1.1.1", nil
}

// DeleteAddress is a silly implementation of the function that deletes an ip on the cloud
func (c *cloudAPI) DeleteAddress(address string) error {
	return nil
}

// Integration contains the silly declarations of all the utilities for the integrations with the cloud
type Integration struct{}

// GetCloudAPI returns a silly cloud api instance according to what is declared in the pool
func (c *Integration) GetCloudAPI(cloudIntegrationOpts *loadbalancing_v1alpha1.CloudIntegrations) clouds.CloudAPI {
	if cloudIntegrationOpts.Hetzner != nil {
		return &cloudAPI{}
	}

	return nil
}

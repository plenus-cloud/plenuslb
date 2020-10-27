package v1alpha1

// PoolOptions is the type for IPPoolSpec options field
type PoolOptions struct {
	HostNetworkInterface *HostNetworkInterfaceOptions `json:"hostNetworkInterface"`
}

// HostNetworkInterfaceOptions are the options required if you wish to add the ips to the interface
type HostNetworkInterfaceOptions struct {
	AddAddressesToInterface bool   `json:"addAddressesToInterface"`
	InterfaceName           string `json:"interfaceName"`
}

// CloudIntegrations is the type for IPPoolSpec cloudIntegration field
type CloudIntegrations struct {
	Hetzner *HetznerCloud `json:"hetzner"`
}

// HetznerCloud is the type for CloudIntegrations hetzner provider
type HetznerCloud struct {
	Token string `json:"token"`
}

// IPPoolStatus is the IPPool status
type IPPoolStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

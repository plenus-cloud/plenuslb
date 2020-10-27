package persistentips

import "plenus.io/plenuslb/pkg/clouds"

var cloudsIntegration clouds.Clouds

// Init performs all the startup operation for persistent ips
func Init() {
	cloudsIntegration = &clouds.Integration{}
	createIPPoolsWatcher()

	warmupIPPoolsCacheOrDie()
	warmupIPAvailabilityOrDie()
}

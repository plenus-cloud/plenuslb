package ephemeralips

import (
	"plenus.io/plenuslb/pkg/clouds"
)

var cloudsIntegration clouds.Clouds

// Init performs all the startup operation for ephemeral ips
func Init() {
	cloudsIntegration = &clouds.Integration{}
	createIPPoolsWatcher()
	warmupIPPoolsCacheOrDie()
}

package clients

import (
	apiextensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	plenuslbclientset "plenus.io/plenuslb/pkg/client/clientset/versioned"
)

var (
	k8sClientset       *clientset.Clientset
	extensionClientset *apiextensionclientset.Clientset
	plenuslbClientset  *plenuslbclientset.Clientset
	config             *rest.Config
)

// GetK8sClient return the k8s basic clientset
var GetK8sClient = func() clientset.Interface {
	return k8sClientset
}

// GetExtensionClient return the clientset for extensions (CRD)
var GetExtensionClient = func() *apiextensionclientset.Clientset {
	return extensionClientset
}

// GetPlenuslbClient return the clientset for plenuslb CRD
var GetPlenuslbClient = func() plenuslbclientset.Interface {
	return plenuslbClientset
}

// CreateClientsOrDie initializes the clients or die if error
func CreateClientsOrDie() *rest.Config {
	var err error
	config, err = rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Cannot create in-cluster config: %s", err.Error())
	}

	k8sClientset = clientset.NewForConfigOrDie(config)
	plenuslbClientset = plenuslbclientset.NewForConfigOrDie(config)
	extensionClientset = apiextensionclientset.NewForConfigOrDie(config)

	return config
}

// GetInClusterConfig return the cluster config
func GetInClusterConfig() *rest.Config {
	return config
}

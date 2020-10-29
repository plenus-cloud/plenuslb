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

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

package servicesupdater

import (
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// UpdateServiceIngressWithIps adds the ip address to the ingresses list of a service
func UpdateServiceIngressWithIps(serviceNamespace, serviceName string, ips []string) error {
	klog.Infof("Updating ingress of service %s/%s with ips %v", serviceNamespace, serviceName, ips)

	k8sClient := clients.GetK8sClient()
	// can't use the store due circular damned dependency
	service, err := k8sClient.CoreV1().Services(serviceNamespace).Get(serviceName, meta_v1.GetOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}

	hasExternalIps, _ := utils.ServiceHasExternalIPs(service)

	if hasExternalIps {
		klog.Infof("Service %s/%s has external ips %v, status should not be updated", serviceNamespace, serviceName, service.Spec.ExternalIPs)
		return nil
	}

	s := service.DeepCopy()
	ingresses := []v1.LoadBalancerIngress{}
	for _, ip := range ips {
		ingresses = append(ingresses, v1.LoadBalancerIngress{IP: ip})
	}
	s.Status.LoadBalancer.Ingress = ingresses

	_, err = k8sClient.CoreV1().Services(serviceNamespace).UpdateStatus(s)
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("Address %v set as ingresses of service %s/%s", ips, serviceNamespace, serviceName)
	return nil
}

// RemoveServiceIngressIPs removes the ip address from the ingresses list of a service
func RemoveServiceIngressIPs(serviceNamespace, serviceName string) error {
	klog.Infof("Removing ingress of service %s/%s", serviceNamespace, serviceName)

	k8sClient := clients.GetK8sClient()
	// can't use the store due circular damned dependency
	service, err := k8sClient.CoreV1().Services(serviceNamespace).Get(serviceName, meta_v1.GetOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}

	s := service.DeepCopy()
	s.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{}

	_, err = k8sClient.CoreV1().Services(serviceNamespace).UpdateStatus(s)
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("Removed ingress addresses fromo service %s/%s", serviceNamespace, serviceName)
	return nil
}

// https://github.com/helm/helm/blob/master/pkg/kube/wait.go

package wait

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// ForResources polls to get the current status of the given Daemonset
// until all are ready or a timeout is reached
func ForResources(kcs clientset.Interface, timeout time.Duration, daemonset *appsv1.DaemonSet) error {
	klog.Infof("beginning wait resources with timeout of %v", timeout)

	return wait.Poll(2*time.Second, timeout, func() (bool, error) {
		pods := []v1.Pod{}

		klog.Info("Checking daemonset resources")
		// check daeonset
		list, err := GetPodsFromDaemonset(kcs, daemonset.Namespace, daemonset.Spec.Selector.MatchLabels)
		if err != nil {
			return false, err
		}
		pods = append(pods, list...)

		isReady := podsReady(pods)
		return isReady, nil
	})
}

// ForCRD polls to get the current status of the given CRD
// until all are ready or a timeout is reached
func ForCRD(timeout time.Duration, crd apiextv1.CustomResourceDefinition) error {
	klog.Infof("beginning wait resources with timeout of %v", timeout)

	extClient := clients.GetExtensionClient()
	return wait.Poll(2*time.Second, timeout, func() (bool, error) {
		klog.Infof("Checking if crd %s is ready", crd.GetName())
		existingResource, err := extClient.ApiextensionsV1().CustomResourceDefinitions().Get(crd.GetName(), metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if !crdReady(*existingResource) {
			return false, nil
		}
		klog.Infof("CRD %s is ready", crd.GetName())
		return true, nil
	})
}

func podsReady(pods []v1.Pod) bool {
	for _, pod := range pods {
		if !utils.IsPodReady(&pod) {
			klog.Infof("Pod is not ready: %s/%s", pod.GetNamespace(), pod.GetName())
			return false
		}
	}
	return true
}

// GetPodsFromDaemonset returns all the pods associated to a give daemonset
func GetPodsFromDaemonset(client kubernetes.Interface, namespace string, selector map[string]string) ([]v1.Pod, error) {
	list, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		FieldSelector: fields.Everything().String(),
		LabelSelector: labels.Set(selector).AsSelector().String(),
	})
	return list.Items, err
}

func crdReady(crd apiextv1.CustomResourceDefinition) bool {
	for _, cond := range crd.Status.Conditions {
		switch cond.Type {
		case apiextv1.Established:
			if cond.Status == apiextv1.ConditionTrue {
				return true
			}
		case apiextv1.NamesAccepted:
			if cond.Status == apiextv1.ConditionFalse {
				// This indicates a naming conflict, but it's probably not the
				// job of this function to fail because of that. Instead,
				// we treat it as a success, since the process should be able to
				// continue.
				return true
			}
		}
	}
	return false
}

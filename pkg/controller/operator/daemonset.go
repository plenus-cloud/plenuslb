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

package operator

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"

	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/events"
	"plenus.io/plenuslb/pkg/controller/utils"
	plwait "plenus.io/plenuslb/pkg/controller/wait"
)

const operatorName = "plenuslb-operator"

const operatorVersion = "v1alpha2"

var (
	deployed        bool
	myNamespace     string
	imagePullSecret string
	image           string
	deployMutex     *sync.Mutex

	stopWatch chan struct{}

	operatorNodesStore cache.Store
	// OperatorNodesController is the operator nodes cache controller
	OperatorNodesController cache.Controller
)

// Init performs all the sturtup tasks
func Init() {
	myNamespace = os.Getenv("MY_POD_NAMESPACE")
	if myNamespace == "" {
		klog.Fatal("MY_POD_NAMESPACE env variable is required!")
	}
	image = os.Getenv("OPERATOR_IMAGE")
	if image == "" {
		klog.Fatal("OPERATOR_IMAGE env variable is required!")
	}
	imagePullSecret = os.Getenv("OPERATOR_PULL_SECRET")
	deployMutex = &sync.Mutex{}

	operatorDaemonset := checkIfOperatorIsDeployedAndUpdatedOrDie()
	if deployed && operatorDaemonset != nil {
		buildOperatorNodesWatcher(operatorDaemonset)
		if err := warmupControllerNodesCacheOrDie(operatorDaemonset); err != nil {
			klog.Fatal(err)
		}
		stopWatch = make(chan struct{})
		WhatchControllerNodes()
	}
}

// IsDeployed return true if the operator is deployed
func IsDeployed() bool {
	return deployed
}

// DeployOrDie deploys the operator. If it fails, the process is terminated
func DeployOrDie() error {
	deployMutex.Lock()
	defer deployMutex.Unlock()

	if deployed {
		return nil
	}
	klog.Info("Deploing daemonset")

	daemonset := getDaemonsetSetDefinition()

	clientset := clients.GetK8sClient()

	daemonsetsClient := clientset.AppsV1().DaemonSets(myNamespace)
	operatorDaemonset, err := daemonsetsClient.Create(daemonset)
	if err != nil {
		klog.Error(err)
		deployed = false
		return err
	}
	klog.Infof("Created daemonset %s, waiting for all pod.", operatorDaemonset.GetObjectMeta().GetName())
	err = plwait.ForResources(clientset, 2*time.Minute, operatorDaemonset)
	if err == nil {
		deployed = true
		buildOperatorNodesWatcher(operatorDaemonset)
		klog.Info("All daemonset's resources are ready, deploy was a succes!")
		stopWatch = make(chan struct{})
		WhatchControllerNodes()
	} else {
		deployed = false
		_ = Delete()
		klog.Error("Something went wrong creating resources for daemonset")
		klog.Fatal(err)
	}
	return err
}

func buildOperatorNodesWatcher(daemonset *appsv1.DaemonSet) {
	klog.Info("Building whatcher for operator nodes")
	selector := daemonset.Spec.Selector.MatchLabels
	optionsModifier := func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Everything().String()
		options.LabelSelector = labels.Set(selector).AsSelector().String()
	}
	watchlist := cache.NewFilteredListWatchFromClient(
		clients.GetK8sClient().CoreV1().RESTClient(),
		string(v1.ResourcePods),
		myNamespace,
		optionsModifier,
	)

	store, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		watchlist,
		&v1.Pod{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}

				klog.Infof("New operator %s", pod.GetName())
				if utils.IsPodReady(pod) {
					newOperatorNode(pod)
				} else {
					klog.Infof("New operator node %s is not ready", pod.GetName())
				}

			},
			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				klog.Infof("Deleted operator node %s", pod.GetName())
				operatorNodeLost(pod)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newPod, ok := newObj.(*v1.Pod)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(newObj))
					return
				}
				oldPod, ok := oldObj.(*v1.Pod)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(oldObj))
					return
				}

				if utils.IsPodReady(newPod) != utils.IsPodReady(oldPod) {
					// now is running
					if utils.IsPodReady(newPod) {
						klog.Infof("Modified operator node %s, pod is now ready", newPod.GetName())
						newOperatorNode(newPod)
					} else {
						klog.Warningf("Modified operator node %s, pod is now not ready", newPod.GetName())
						operatorNodeLost(oldPod)
					}

				}
			},
		},
	)

	operatorNodesStore = store
	OperatorNodesController = controller
}

// WhatchControllerNodes start the whtcher for the nodes of the controller
func WhatchControllerNodes() {
	go OperatorNodesController.Run(stopWatch)
}

func warmupControllerNodesCacheOrDie(daemonset *appsv1.DaemonSet) error {
	list, err := plwait.GetPodsFromDaemonset(clients.GetK8sClient(), daemonset.Namespace, daemonset.Spec.Selector.MatchLabels)
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("Warming-up operators nodes cache, %d nodes are on the cluster now", len(list))

	for _, pod := range list {
		if utils.IsPodReady(&pod) {
			klog.Infof("Operator node %s is running on cluster node %s", pod.GetName(), pod.Spec.NodeName)
			if err := operatorNodesStore.Add(pod); err != nil {
				klog.Error(err)
				// TODO: return err
				// return err
			}
		} else {
			klog.Warningf("Operator node %s is not ready, is %s for the following reason: %s", pod.GetName(), pod.Status.Phase, pod.Status.Reason)
		}
	}
	return nil
}

func newOperatorNode(pod *v1.Pod) {
	events.NewOperatorNode(pod.Spec.NodeName)
}

func operatorNodeLost(pod *v1.Pod) {
	events.OperatorNodeLost(pod.Spec.NodeName)
	checkIfOperatorIsStillRequiredOrDie()
}

func checkIfOperatorIsStillRequiredOrDie() {
	// if all the operator nodes are gone, check if is still required
	// if yes, try to deploy it again or die
	if GetOperatorsCount() == 0 && deployed {
		klog.Info("All operators lost but operator is required at least by pool, checking status again")
		operator := checkIfOperatorIsDeployedAndUpdatedOrDie()
		if operator == nil {
			klog.Warning("Operator is not deployed, trying to deploy again")
			err := DeployOrDie()
			if err != nil {
				klog.Error("Failed to deploy operator. Please, don't touch my operator")
			}
		}
	}
}

func getDaemonsetSetDefinition() *appsv1.DaemonSet {
	imagePullSecrets := []v1.LocalObjectReference{}
	if imagePullSecret != "" {
		imagePullSecrets = []v1.LocalObjectReference{{
			Name: imagePullSecret,
		}}
	}
	privileged := true
	tolerationSeconds := int64(2)
	healthPort := utils.HealthPort()
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operatorName,
			Namespace: myNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "plenuslb-operator",
				"app.kubernetes.io/name":       "plenuslb-operator",
				"app.kubernetes.io/managed-by": "plenuslb",
				"app.kubernetes.io/version":    operatorVersion,
			},
		},

		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/instance": "plenuslb-operator",
					"app.kubernetes.io/name":     "plenuslb-operator",
				},
			},

			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/instance": "plenuslb-operator",
						"app.kubernetes.io/name":     "plenuslb-operator",
					},
				},
				Spec: v1.PodSpec{
					ImagePullSecrets: imagePullSecrets,
					HostNetwork:      true,
					Containers: []v1.Container{
						{
							Name:            operatorName,
							Image:           image,
							ImagePullPolicy: v1.PullAlways,
							SecurityContext: &v1.SecurityContext{
								Privileged: &privileged,
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "health",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: healthPort,
								},
								{
									Name:          "grpc",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: OperatorPort,
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/health",
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "health",
										},
									},
								},
								InitialDelaySeconds: 2,
								PeriodSeconds:       1,
								FailureThreshold:    10,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/health",
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "health",
										},
									},
								},
								InitialDelaySeconds: 2,
								PeriodSeconds:       1,
								FailureThreshold:    5,
							},
							Env: []v1.EnvVar{
								{
									Name: "MY_NODE_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name:  "HEALTH_PORT",
									Value: fmt.Sprintf("%d", healthPort),
								},
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("50m"),
									v1.ResourceMemory: resource.MustParse("100Mi"),
								},
							},
						},
					},

					Tolerations: []v1.Toleration{
						{
							Key:               "node.kubernetes.io/not-ready",
							Effect:            v1.TaintEffectNoExecute,
							Operator:          v1.TolerationOpExists,
							TolerationSeconds: &tolerationSeconds,
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Effect:            v1.TaintEffectNoExecute,
							Operator:          v1.TolerationOpExists,
							TolerationSeconds: &tolerationSeconds,
						},
					},
				},
			},
		},
	}
}

// Delete deletes the operators
func Delete() error {

	klog.Info("Deleting daemonset")
	daemonsetsClient := clients.GetK8sClient().AppsV1().DaemonSets(myNamespace)
	propagation := metav1.DeletePropagationForeground
	err := daemonsetsClient.Delete(operatorName, &metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
	if err != nil {
		klog.Error(err)
	} else {
		deployed = false
		close(stopWatch)
		klog.Info("Daemonset deleted")
	}
	return err
}

func checkIfOperatorIsDeployedAndUpdatedOrDie() *appsv1.DaemonSet {
	klog.Info("Checking if operator is already present and up-to-date")

	daemonsetsClient := clients.GetK8sClient().AppsV1().DaemonSets(myNamespace)
	daemonset, err := daemonsetsClient.Get(operatorName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		deployed = false
		klog.Info("Operator is not present")
		return nil
	} else if err == nil {
		if checkIfOperatorIsUpdated(daemonset) {
			klog.Info("Operator is already present and up-to-date")
			deployed = true
			return daemonset
		}
		klog.Info("Updating operator")
		if err := updateOperator(); err != nil {
			klog.Fatal(err)
			return nil
		}
		klog.Info("Operator have been up-to-dated")
		deployed = true
		return daemonset

	} else {
		klog.Fatal(err)
	}
	return nil
}

func checkIfOperatorIsUpdated(daemonset *appsv1.DaemonSet) bool {
	isUpdated := true
	currentVersion := daemonset.GetLabels()["app.kubernetes.io/version"]
	if currentVersion != operatorVersion {
		isUpdated = false
		klog.Infof("Operator is not updated, version is %s instead of %s", currentVersion, operatorVersion)
	}

	currentImage := daemonset.Spec.Template.Spec.Containers[0].Image
	if currentImage != image {
		isUpdated = false
		klog.Infof("Operator is not updated, image is %s instead of %s", currentImage, image)
	}

	return isUpdated
}

func updateOperator() error {

	daemonsetsClient := clients.GetK8sClient().AppsV1().DaemonSets(myNamespace)

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		daemonset, err := daemonsetsClient.Get(operatorName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get latest version of operator: %v", err)
			return err
		}
		_, updateErr := daemonsetsClient.Update(mergeOperatorsDefinitions(daemonset))
		return updateErr
	})
	if retryErr != nil {
		klog.Errorf("Update failed: %v", retryErr)
		return retryErr
	}

	return nil
}

func mergeOperatorsDefinitions(oldOperator *appsv1.DaemonSet) *appsv1.DaemonSet {
	newOperator := getDaemonsetSetDefinition()
	newOperator.ResourceVersion = oldOperator.ResourceVersion
	newOperator.ObjectMeta.ResourceVersion = oldOperator.ObjectMeta.ResourceVersion
	return newOperator
}

// GetStoreList returns the store af the operator nodes
var GetStoreList = func() []interface{} {
	return operatorNodesStore.List()
}

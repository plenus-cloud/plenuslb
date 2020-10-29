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

package servicewatcher

import (
	"errors"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	allocationreconciler "plenus.io/plenuslb/pkg/controller/allocationReconciler"
	allocationslock "plenus.io/plenuslb/pkg/controller/allocationsLock"
	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/ipallocations"
	"plenus.io/plenuslb/pkg/controller/utils"
)

var (
	servicesStore cache.Store
	// ServicesController manages the services resources
	ServicesController cache.Controller
)

// ErrServiceNotFound returned when the requested service deas not exists
var ErrServiceNotFound = errors.New("Service not found")

// Init initializes performs all the startup tasks for the services whatcher
func Init() {
	buildServicesWatcher()
}

// buildServicesWatcher build the whatcher for load balancer services
func buildServicesWatcher() {
	watchlist := cache.NewListWatchFromClient(
		clients.GetK8sClient().CoreV1().RESTClient(),
		string(v1.ResourceServices),
		v1.NamespaceAll,
		fields.Everything(),
	)
	store, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		watchlist,
		&v1.Service{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				service, ok := obj.(*v1.Service)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				serviceCreated(service)
			},
			DeleteFunc: func(obj interface{}) {
				service, ok := obj.(*v1.Service)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
					return
				}
				serviceDeleted(service)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newService, ok := newObj.(*v1.Service)
				if !ok {
					klog.Errorf("unexpected type %s", reflect.TypeOf(newObj))
					return
				}
				serviceChanged(newService)
			},
		},
	)

	servicesStore = store
	ServicesController = controller
}

// WatchServices watch all k8s services and manages those LoadBalancers
func WatchServices(stop chan struct{}) {
	go ServicesController.Run(stop)
}

func serviceCreated(service *v1.Service) {
	if utils.ServiceIsLoadBalancer(service) {
		klog.Infof("Added new LoadBalancer service %s/%s", service.GetNamespace(), service.GetName())
		go allocationreconciler.CreateAllocationForService(service)
	}
}

func serviceDeleted(service *v1.Service) {
	if utils.ServiceIsLoadBalancer(service) {
		klog.Infof("Deleted LoadBalancer service %s", service.GetName())
		if err := ipallocations.DeleteAllocationByName(service.GetNamespace(), service.GetName()); err != nil {
			klog.Error(err)
		}
	}
}

// is there allocation for this service?
// -> if not create
// -> if yes: check if the allocation is right for the service
func serviceChanged(service *v1.Service) {
	allocation, err := ipallocations.FindAllocation(service.GetNamespace(), service.GetName())
	if err != nil {
		klog.Error(err)
		return
	}
	if allocation == nil && utils.ServiceIsLoadBalancer(service) {
		go allocationreconciler.CreateAllocationForService(service)
	} else if allocation != nil {
		if err := allocationslock.AcquireAllocationLock(allocation); err != nil {
			return
		}
		defer allocationslock.RemoveFromLock(allocation)
		go allocationreconciler.ReconcileAllocation(service, allocation)
	}
}

// FindService returns the requested service by namespace and name
func FindService(serviceNamespace, serviceName string) (*v1.Service, error) {
	if ServicesController.HasSynced() {
		for _, obj := range servicesStore.List() {
			service, ok := obj.(*v1.Service)
			if !ok {
				err := fmt.Errorf("unexpected type %s", reflect.TypeOf(obj))
				klog.Error(err)
				return nil, err
			}
			if service.GetNamespace() == serviceNamespace && service.GetName() == serviceName {
				return service, nil
			}

		}
		return nil, ErrServiceNotFound
	}

	service, err := clients.GetK8sClient().CoreV1().Services(serviceNamespace).Get(serviceName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		klog.Warningf("Service %s/%s not found", serviceNamespace, serviceName)
		return nil, ErrServiceNotFound
	} else if err != nil {
		return nil, err
	}

	return service, nil
}

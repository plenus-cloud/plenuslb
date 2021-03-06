/*
Copyright The Kubernetes Authors.

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
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

// IPAllocationLister helps list IPAllocations.
type IPAllocationLister interface {
	// List lists all IPAllocations in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.IPAllocation, err error)
	// IPAllocations returns an object that can list and get IPAllocations.
	IPAllocations(namespace string) IPAllocationNamespaceLister
	IPAllocationListerExpansion
}

// iPAllocationLister implements the IPAllocationLister interface.
type iPAllocationLister struct {
	indexer cache.Indexer
}

// NewIPAllocationLister returns a new IPAllocationLister.
func NewIPAllocationLister(indexer cache.Indexer) IPAllocationLister {
	return &iPAllocationLister{indexer: indexer}
}

// List lists all IPAllocations in the indexer.
func (s *iPAllocationLister) List(selector labels.Selector) (ret []*v1alpha1.IPAllocation, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.IPAllocation))
	})
	return ret, err
}

// IPAllocations returns an object that can list and get IPAllocations.
func (s *iPAllocationLister) IPAllocations(namespace string) IPAllocationNamespaceLister {
	return iPAllocationNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// IPAllocationNamespaceLister helps list and get IPAllocations.
type IPAllocationNamespaceLister interface {
	// List lists all IPAllocations in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.IPAllocation, err error)
	// Get retrieves the IPAllocation from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.IPAllocation, error)
	IPAllocationNamespaceListerExpansion
}

// iPAllocationNamespaceLister implements the IPAllocationNamespaceLister
// interface.
type iPAllocationNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all IPAllocations in the indexer for a given namespace.
func (s iPAllocationNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.IPAllocation, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.IPAllocation))
	})
	return ret, err
}

// Get retrieves the IPAllocation from the indexer for a given namespace and name.
func (s iPAllocationNamespaceLister) Get(name string) (*v1alpha1.IPAllocation, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("ipallocation"), name)
	}
	return obj.(*v1alpha1.IPAllocation), nil
}

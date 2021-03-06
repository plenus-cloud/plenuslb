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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

// FakePersistentIPPools implements PersistentIPPoolInterface
type FakePersistentIPPools struct {
	Fake *FakeLoadbalancingV1alpha1
}

var persistentippoolsResource = schema.GroupVersionResource{Group: "loadbalancing.plenus.io", Version: "v1alpha1", Resource: "persistentippools"}

var persistentippoolsKind = schema.GroupVersionKind{Group: "loadbalancing.plenus.io", Version: "v1alpha1", Kind: "PersistentIPPool"}

// Get takes name of the persistentIPPool, and returns the corresponding persistentIPPool object, and an error if there is any.
func (c *FakePersistentIPPools) Get(name string, options v1.GetOptions) (result *v1alpha1.PersistentIPPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(persistentippoolsResource, name), &v1alpha1.PersistentIPPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PersistentIPPool), err
}

// List takes label and field selectors, and returns the list of PersistentIPPools that match those selectors.
func (c *FakePersistentIPPools) List(opts v1.ListOptions) (result *v1alpha1.PersistentIPPoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(persistentippoolsResource, persistentippoolsKind, opts), &v1alpha1.PersistentIPPoolList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PersistentIPPoolList{ListMeta: obj.(*v1alpha1.PersistentIPPoolList).ListMeta}
	for _, item := range obj.(*v1alpha1.PersistentIPPoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested persistentIPPools.
func (c *FakePersistentIPPools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(persistentippoolsResource, opts))
}

// Create takes the representation of a persistentIPPool and creates it.  Returns the server's representation of the persistentIPPool, and an error, if there is any.
func (c *FakePersistentIPPools) Create(persistentIPPool *v1alpha1.PersistentIPPool) (result *v1alpha1.PersistentIPPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(persistentippoolsResource, persistentIPPool), &v1alpha1.PersistentIPPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PersistentIPPool), err
}

// Update takes the representation of a persistentIPPool and updates it. Returns the server's representation of the persistentIPPool, and an error, if there is any.
func (c *FakePersistentIPPools) Update(persistentIPPool *v1alpha1.PersistentIPPool) (result *v1alpha1.PersistentIPPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(persistentippoolsResource, persistentIPPool), &v1alpha1.PersistentIPPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PersistentIPPool), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePersistentIPPools) UpdateStatus(persistentIPPool *v1alpha1.PersistentIPPool) (*v1alpha1.PersistentIPPool, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(persistentippoolsResource, "status", persistentIPPool), &v1alpha1.PersistentIPPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PersistentIPPool), err
}

// Delete takes name of the persistentIPPool and deletes it. Returns an error if one occurs.
func (c *FakePersistentIPPools) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(persistentippoolsResource, name), &v1alpha1.PersistentIPPool{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePersistentIPPools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(persistentippoolsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.PersistentIPPoolList{})
	return err
}

// Patch applies the patch and returns the patched persistentIPPool.
func (c *FakePersistentIPPools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PersistentIPPool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(persistentippoolsResource, name, pt, data, subresources...), &v1alpha1.PersistentIPPool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PersistentIPPool), err
}

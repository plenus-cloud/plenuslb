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

package persistentips

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	fakeclietset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	plenuslbclientset "plenus.io/plenuslb/pkg/client/clientset/versioned"
	plenuslbclientsetfake "plenus.io/plenuslb/pkg/client/clientset/versioned/fake"
	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/operator"
)

var (
	node = v1.Node{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "fakeNodeName",
		},
	}

	poolWithCloudAddresses = []string{"1.1.1.1", "2.2.2.2"}
	poolWithCloud          = loadbalancing_v1alpha1.PersistentIPPool{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test_persistent",
		},
		Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
			Addresses: []string{},
			Options: &loadbalancing_v1alpha1.PoolOptions{
				HostNetworkInterface: &loadbalancing_v1alpha1.HostNetworkInterfaceOptions{
					AddAddressesToInterface: true,
					InterfaceName:           "test",
				},
			},
			CloudIntegration: &loadbalancing_v1alpha1.CloudIntegrations{
				Hetzner: &loadbalancing_v1alpha1.HetznerCloud{
					Token: "fake_token",
				},
			},
		},
	}

	poolWithoutCloudAddresses = []string{"5.1.1.1", "5.2.2.2"}
	poolWithoutCloud          = loadbalancing_v1alpha1.PersistentIPPool{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test_persistent_no_cloud",
		},
		Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
			Addresses: []string{},
			Options: &loadbalancing_v1alpha1.PoolOptions{
				HostNetworkInterface: &loadbalancing_v1alpha1.HostNetworkInterfaceOptions{
					AddAddressesToInterface: true,
					InterfaceName:           "test",
				},
			},
		},
	}

	poolWithoutHostAddresses = []string{"20.1.1.1", "20.2.2.2"}
	poolWithoutHost          = loadbalancing_v1alpha1.PersistentIPPool{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test_persistent_no_host",
		},
		Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
			Addresses: []string{},
			CloudIntegration: &loadbalancing_v1alpha1.CloudIntegrations{
				Hetzner: &loadbalancing_v1alpha1.HetznerCloud{
					Token: "fake_token",
				},
			},
		},
	}

	poolWithoutCloudAndHostAddresses = []string{"10.1.1.1", "10.2.2.2"}
	poolWithoutCloudAndHost          = loadbalancing_v1alpha1.PersistentIPPool{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test_persistent_no_cloud_no_host",
		},
		Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
			Addresses: []string{},
		},
	}

	operatorPod = v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "fakePod",
		},
		Spec: v1.PodSpec{
			NodeName: "fakePodNodeName",
		},
		Status: v1.PodStatus{
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}
)

func buildPoolsMock() {
	poolWithCloud.Spec.Addresses = append(poolWithCloud.Spec.Addresses, poolWithCloudAddresses...)

	poolWithoutCloud.Spec.Addresses = append(poolWithoutCloud.Spec.Addresses, poolWithoutCloudAddresses...)

	poolWithoutHost.Spec.Addresses = append(poolWithoutHost.Spec.Addresses, poolWithoutHostAddresses...)

	poolWithoutCloudAndHost.Spec.Addresses = append(poolWithoutCloudAndHost.Spec.Addresses, poolWithoutCloudAndHostAddresses...)

	mockPersistentPoolCache(&poolWithCloud, &poolWithoutCloud, &poolWithoutHost, &poolWithoutCloudAndHost)
}

func restorePoolsAvailability() {
	UpdatePoolAvailability(&poolWithCloud, poolWithCloudAddresses)
	UpdatePoolAvailability(&poolWithoutCloud, poolWithoutCloudAddresses)
	UpdatePoolAvailability(&poolWithoutHost, poolWithoutHostAddresses)
	UpdatePoolAvailability(&poolWithoutCloudAndHost, poolWithoutCloudAndHostAddresses)
}

func mockGetPlenuslbClient(objects ...runtime.Object) {
	clients.GetPlenuslbClient = func() plenuslbclientset.Interface {
		return plenuslbclientsetfake.NewSimpleClientset(objects...)
	}
}

func mockGetK8sClient(objects ...runtime.Object) {
	clients.GetK8sClient = func() clientset.Interface {
		return fakeclietset.NewSimpleClientset(objects...)
	}
}

func mockCoreClientCache(pods ...*v1.Pod) {
	testStoreKeyFunc := func(obj interface{}) (string, error) {
		return fmt.Sprintf("%s/%s", obj.(*v1.Pod).Name, obj.(*v1.Pod).Namespace), nil
	}
	newStore := cache.NewStore(testStoreKeyFunc)

	for _, pod := range pods {
		_ = newStore.Add(pod)
	}
	operator.GetStoreList = func() []interface{} {
		return newStore.List()
	}
}

func mockPersistentPoolCache(pools ...*loadbalancing_v1alpha1.PersistentIPPool) {
	testStoreKeyFunc := func(obj interface{}) (string, error) {
		return obj.(*loadbalancing_v1alpha1.PersistentIPPool).Name, nil
	}
	newStore := cache.NewStore(testStoreKeyFunc)

	for _, pool := range pools {
		_ = newStore.Add(pool)
	}
	poolStoreList = func() []interface{} {
		return newStore.List()
	}
}

func Test_buildAllocations(t *testing.T) {
	buildPoolsMock()
	mockGetK8sClient(node.DeepCopyObject())
	mockCoreClientCache(&operatorPod)

	type args struct {
		serviceNamespace string
		ips              []string
	}
	tests := []struct {
		name    string
		args    args
		want    []*loadbalancing_v1alpha1.IPAllocationAddresses
		wantErr bool
	}{
		{
			name: "Should fail due IP not available",
			args: args{
				serviceNamespace: "fake_ns",
				ips:              []string{"100.1.1.1"},
			},
			want:    []*loadbalancing_v1alpha1.IPAllocationAddresses{},
			wantErr: true,
		},
		{
			name: "should build",
			args: args{
				serviceNamespace: "fake_ns",
				ips:              []string{"1.1.1.1", "5.1.1.1", "10.1.1.1", "20.1.1.1"},
			},
			want: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address:          "1.1.1.1",
					Pool:             "test_persistent",
					NetworkInterface: "test",
					CloudProvider:    "hetzner",
					NodeName:         "fakePodNodeName",
				},
				{
					Address:          "5.1.1.1",
					Pool:             "test_persistent_no_cloud",
					NetworkInterface: "test",
					NodeName:         "fakePodNodeName",
				},
				{
					Address: "10.1.1.1",
					Pool:    "test_persistent_no_cloud_no_host",
				},
				{
					Address:       "20.1.1.1",
					Pool:          "test_persistent_no_host",
					CloudProvider: "hetzner",
					NodeName:      "fakeNodeName",
				},
			},
		},
	}
	for _, tt := range tests {
		restorePoolsAvailability()
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAllocations(tt.args.serviceNamespace, nil, tt.args.ips)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildAllocations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildAllocations() = %v, want %v", allocationArrayToString(got), allocationArrayToString(tt.want))
			}
		})
	}
}

func allocationArrayToString(allocations []*loadbalancing_v1alpha1.IPAllocationAddresses) string {
	r := "["
	for _, al := range allocations {
		r = fmt.Sprintf("%s, %v", r, *al)
	}
	r = fmt.Sprintf("%s]", r)
	return r
}

func Test_patchPersistentAllocation(t *testing.T) {
	buildPoolsMock()
	mockGetK8sClient(node.DeepCopyObject())
	mockCoreClientCache(&operatorPod)

	serviceNamespace := "serviceNamespace"
	serviceName := "serviceName"

	allocWithAllIPS := &loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
		},
		Spec: loadbalancing_v1alpha1.IPAllocationSpec{
			Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address:          "1.1.1.1",
					Pool:             "test_persistent",
					NetworkInterface: "test",
					CloudProvider:    "hetzner",
					NodeName:         "fakePodNodeName",
				},
				{
					Address:          "5.1.1.1",
					Pool:             "test_persistent_no_cloud",
					NetworkInterface: "test",
					NodeName:         "fakePodNodeName",
				},
				{
					Address: "10.1.1.1",
					Pool:    "test_persistent_no_cloud_no_host",
				},
				{
					Address:       "20.1.1.1",
					Pool:          "test_persistent_no_host",
					CloudProvider: "hetzner",
					NodeName:      "fakeNodeName",
				},
			},
		},
	}

	mockGetPlenuslbClient(allocWithAllIPS.DeepCopyObject())

	waitingStatus := loadbalancing_v1alpha1.IPAllocationStatus{
		State:   loadbalancing_v1alpha1.AllocationStatusPending,
		Message: "Waiting for allocator",
	}

	noIPFailedStatus := loadbalancing_v1alpha1.IPAllocationStatus{
		State:   loadbalancing_v1alpha1.AllocationStatusError,
		Message: "No ip available",
	}
	wantWithAllIPS := allocWithAllIPS.DeepCopy()
	wantWithAllIPS.Status = waitingStatus

	type args struct {
		serviceNamespace   string
		serviceName        string
		ips                []string
		actualAllocationRO *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name                   string
		args                   args
		wantRemovedAllocations []*loadbalancing_v1alpha1.IPAllocationAddresses
		want                   *loadbalancing_v1alpha1.IPAllocation
		wantErr                bool
		wantAllocationErr      bool
	}{
		{
			name: "should be the same",
			args: args{
				serviceNamespace:   serviceNamespace,
				serviceName:        serviceName,
				ips:                []string{"1.1.1.1", "5.1.1.1", "10.1.1.1", "20.1.1.1"},
				actualAllocationRO: allocWithAllIPS.DeepCopy(),
			},
			want: allocWithAllIPS,
		},
		{
			name: "should remove one",
			args: args{
				serviceNamespace:   serviceNamespace,
				serviceName:        serviceName,
				ips:                []string{"1.1.1.1", "5.1.1.1", "10.1.1.1"},
				actualAllocationRO: allocWithAllIPS.DeepCopy(),
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      serviceName,
					Namespace: serviceNamespace,
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.1",
							Pool:             "test_persistent",
							NetworkInterface: "test",
							CloudProvider:    "hetzner",
							NodeName:         "fakePodNodeName",
						},
						{
							Address:          "5.1.1.1",
							Pool:             "test_persistent_no_cloud",
							NetworkInterface: "test",
							NodeName:         "fakePodNodeName",
						},
						{
							Address: "10.1.1.1",
							Pool:    "test_persistent_no_cloud_no_host",
						},
					},
				},
				Status: waitingStatus,
			},
			wantRemovedAllocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address:       "20.1.1.1",
					Pool:          "test_persistent_no_host",
					CloudProvider: "hetzner",
					NodeName:      "fakeNodeName",
				},
			},
		},
		{
			name: "should remove two",
			args: args{
				serviceNamespace:   serviceNamespace,
				serviceName:        serviceName,
				ips:                []string{"1.1.1.1", "5.1.1.1"},
				actualAllocationRO: allocWithAllIPS.DeepCopy(),
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      serviceName,
					Namespace: serviceNamespace,
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.1",
							Pool:             "test_persistent",
							NetworkInterface: "test",
							CloudProvider:    "hetzner",
							NodeName:         "fakePodNodeName",
						},
						{
							Address:          "5.1.1.1",
							Pool:             "test_persistent_no_cloud",
							NetworkInterface: "test",
							NodeName:         "fakePodNodeName",
						},
					},
				},
				Status: waitingStatus,
			},
			wantRemovedAllocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address: "10.1.1.1",
					Pool:    "test_persistent_no_cloud_no_host",
				},
				{
					Address:       "20.1.1.1",
					Pool:          "test_persistent_no_host",
					CloudProvider: "hetzner",
					NodeName:      "fakeNodeName",
				},
			},
		},
		{
			name: "should add one",
			args: args{
				serviceNamespace:   serviceNamespace,
				serviceName:        serviceName,
				ips:                []string{"1.1.1.1", "5.1.1.1", "10.1.1.1", "20.1.1.1", "20.2.2.2"},
				actualAllocationRO: allocWithAllIPS.DeepCopy(),
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      serviceName,
					Namespace: serviceNamespace,
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.1",
							Pool:             "test_persistent",
							NetworkInterface: "test",
							CloudProvider:    "hetzner",
							NodeName:         "fakePodNodeName",
						},
						{
							Address:          "5.1.1.1",
							Pool:             "test_persistent_no_cloud",
							NetworkInterface: "test",
							NodeName:         "fakePodNodeName",
						},
						{
							Address: "10.1.1.1",
							Pool:    "test_persistent_no_cloud_no_host",
						},
						{
							Address:       "20.1.1.1",
							Pool:          "test_persistent_no_host",
							CloudProvider: "hetzner",
							NodeName:      "fakeNodeName",
						},
						{
							Address:       "20.2.2.2",
							Pool:          "test_persistent_no_host",
							CloudProvider: "hetzner",
							NodeName:      "fakeNodeName",
						},
					},
				},
				Status: waitingStatus,
			},
		},
		{
			name: "should fail due no available ip",
			args: args{
				serviceNamespace:   serviceNamespace,
				serviceName:        serviceName,
				ips:                []string{"1.1.1.1", "5.1.1.1", "10.1.1.1", "20.1.1.1", "20.2.2.5"},
				actualAllocationRO: allocWithAllIPS.DeepCopy(),
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      serviceName,
					Namespace: serviceNamespace,
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.1",
							Pool:             "test_persistent",
							NetworkInterface: "test",
							CloudProvider:    "hetzner",
							NodeName:         "fakePodNodeName",
						},
						{
							Address:          "5.1.1.1",
							Pool:             "test_persistent_no_cloud",
							NetworkInterface: "test",
							NodeName:         "fakePodNodeName",
						},
						{
							Address: "10.1.1.1",
							Pool:    "test_persistent_no_cloud_no_host",
						},
						{
							Address:       "20.1.1.1",
							Pool:          "test_persistent_no_host",
							CloudProvider: "hetzner",
							NodeName:      "fakeNodeName",
						},
					},
				},
				Status: noIPFailedStatus,
			},
			wantAllocationErr: true,
		},
	}

	for _, tt := range tests {
		restorePoolsAvailability()
		t.Run(tt.name, func(t *testing.T) {
			got, wantRemovedAllocations, allocationErr, err := CheckAndPatchAllocation(tt.args.serviceNamespace, tt.args.serviceName, tt.args.ips, tt.args.actualAllocationRO)
			if (err != nil) != tt.wantErr {
				t.Errorf("patchPersistentAllocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (allocationErr != nil) != tt.wantAllocationErr {
				t.Errorf("patchPersistentAllocation() allocationErr = %v, wantAllocationErr %v", err, tt.wantErr)
				return
			}
			if (got != nil) != (tt.want != nil) ||
				got != nil &&
					(!assert.Equal(t, tt.want.Status, got.Status) ||
						!assert.Equal(t, tt.want.ObjectMeta, got.ObjectMeta) ||
						got.Spec.Type != tt.want.Spec.Type ||
						!assert.ElementsMatch(t, tt.want.Spec.Allocations, got.Spec.Allocations)) {
				t.Errorf("patchPersistentAllocation() = %v (status: %v), want %v (status: %v)", allocationArrayToString(got.Spec.Allocations), got.Status, allocationArrayToString(tt.want.Spec.Allocations), tt.want.Status)
			}
			if !assert.ElementsMatch(t, wantRemovedAllocations, tt.wantRemovedAllocations) {
				t.Errorf("patchPersistentAllocation() removedAllocations = %v, want %v", allocationArrayToString(wantRemovedAllocations), allocationArrayToString(tt.wantRemovedAllocations))
			}
		})
	}
}

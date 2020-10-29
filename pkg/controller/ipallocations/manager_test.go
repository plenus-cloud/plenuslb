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

package ipallocations

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	plenuslbclientset "plenus.io/plenuslb/pkg/client/clientset/versioned"
	plenuslbclientsetfake "plenus.io/plenuslb/pkg/client/clientset/versioned/fake"
	"plenus.io/plenuslb/pkg/controller/clients"
)

func mockGetPlenuslbClient(objects ...runtime.Object) {
	clients.GetPlenuslbClient = func() plenuslbclientset.Interface {
		return plenuslbclientsetfake.NewSimpleClientset(objects...)
	}
}

func TestCreateAllocation(t *testing.T) {

	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_name_exists",
			Namespace: "test_namespace",
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	type args struct {
		namespace      string
		name           string
		allocationType loadbalancing_v1alpha1.IPType
		allocations    []*loadbalancing_v1alpha1.IPAllocationAddresses
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should create",
			args: args{
				namespace:      "test_namespace",
				name:           "test_name",
				allocationType: loadbalancing_v1alpha1.EphemeralIP,
				allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
					{
						Address:          "1.1.1.1",
						NetworkInterface: "fake",
						NodeName:         "fake",
						CloudProvider:    "fake",
						Pool:             "fake",
					},
				},
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "plenuslb",
					},
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Type: loadbalancing_v1alpha1.EphemeralIP,
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.1",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
					},
				},
			},
		},
		{
			name: "should fail, only one allocation for ephemeral",
			args: args{
				namespace:      "test_namespace",
				name:           "test_name",
				allocationType: loadbalancing_v1alpha1.EphemeralIP,
				allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
					{
						Address:          "1.1.1.1",
						NetworkInterface: "fake",
						NodeName:         "fake",
						CloudProvider:    "fake",
						Pool:             "fake",
					},
					{
						Address:          "1.1.1.2",
						NetworkInterface: "fake",
						NodeName:         "fake",
						CloudProvider:    "fake",
						Pool:             "fake",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should fail, already exists",
			args: args{
				namespace:      "test_namespace",
				name:           "test_name_exists",
				allocationType: loadbalancing_v1alpha1.EphemeralIP,
				allocations:    []*loadbalancing_v1alpha1.IPAllocationAddresses{},
			},
			wantErr: true,
		},
		{
			name: "baz",
			args: args{
				namespace:      "test_namespace",
				name:           "test_name",
				allocationType: loadbalancing_v1alpha1.PersistentIP,
				allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
					{
						Address:          "1.1.1.1",
						NetworkInterface: "fake",
						NodeName:         "fake",
						CloudProvider:    "fake",
						Pool:             "fake",
					},
					{
						Address:          "1.1.1.2",
						NetworkInterface: "fake",
						NodeName:         "fake",
						CloudProvider:    "fake",
						Pool:             "fake",
					},
				},
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "plenuslb",
					},
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Type: loadbalancing_v1alpha1.PersistentIP,
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.1",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
						{
							Address:          "1.1.1.2",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateAllocation(tt.args.namespace, tt.args.name, tt.args.allocationType, tt.args.allocations)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAllocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAllocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAllocationStatusSuccess(t *testing.T) {
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_name",
			Namespace: "test_namespace",
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	type args struct {
		allocationRO *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should update",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name",
						Namespace: "test_namespace",
					},
				},
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusSuccess,
					Message: "Allocated",
				},
			},
		},
		{
			name: "should fail with not found error",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_wrong",
						Namespace: "test_namespace",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetAllocationStatusSuccess(tt.args.allocationRO)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAllocationStatusSuccess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetAllocationStatusSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAllocationStatusError(t *testing.T) {
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_name",
			Namespace: "test_namespace",
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	reason := errors.New("Reason for the error")
	type args struct {
		allocationRO *loadbalancing_v1alpha1.IPAllocation
		reason       error
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should update",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name",
						Namespace: "test_namespace",
					},
				},
				reason: reason,
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusError,
					Message: reason.Error(),
				},
			},
		},
		{
			name: "should fail with not found error",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_wrong",
						Namespace: "test_namespace",
					},
				},
				reason: reason,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetAllocationStatusError(tt.args.allocationRO, tt.args.reason)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAllocationStatusError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetAllocationStatusError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAllocationStatusNodeError(t *testing.T) {
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_name",
			Namespace: "test_namespace",
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	reason := errors.New("Reason for the error")
	type args struct {
		allocationRO *loadbalancing_v1alpha1.IPAllocation
		reason       error
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should update",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name",
						Namespace: "test_namespace",
					},
				},
				reason: reason,
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusNodeError,
					Message: reason.Error(),
				},
			},
		},
		{
			name: "should fail with not found error",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_wrong",
						Namespace: "test_namespace",
					},
				},
				reason: reason,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetAllocationStatusNodeError(tt.args.allocationRO, tt.args.reason)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAllocationStatusNodeError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetAllocationStatusNodeError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAllocationStatusAddrDeleted(t *testing.T) {
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_name",
			Namespace: "test_namespace",
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	addr := "1.1.1.1"
	type args struct {
		allocationRO *loadbalancing_v1alpha1.IPAllocation
		addr         string
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should update",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name",
						Namespace: "test_namespace",
					},
				},
				addr: addr,
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusAddrDeleted,
					Message: fmt.Sprintf("Address %s removed from pool", addr),
				},
			},
		},
		{
			name: "should fail with not found error",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_wrong",
						Namespace: "test_namespace",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetAllocationStatusAddrDeleted(tt.args.allocationRO, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAllocationStatusAddrDeleted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetAllocationStatusAddrDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAllocationStatusPending(t *testing.T) {
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_name",
			Namespace: "test_namespace",
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	msg := "Waiting for allocator"
	type args struct {
		allocationRO *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should update",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name",
						Namespace: "test_namespace",
					},
				},
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusPending,
					Message: msg,
				},
			},
		},
		{
			name: "should fail with not found error",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_wrong",
						Namespace: "test_namespace",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetAllocationStatusPending(tt.args.allocationRO)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAllocationStatusPending() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetAllocationStatusPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddressRemovedFromAllocation(t *testing.T) {
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_name",
			Namespace: "test_namespace",
		},
		Spec: loadbalancing_v1alpha1.IPAllocationSpec{
			Type: loadbalancing_v1alpha1.PersistentIP,
			Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address:          "1.1.1.1",
					NetworkInterface: "fake",
					NodeName:         "fake",
					CloudProvider:    "fake",
					Pool:             "fake",
				},
				{
					Address:          "1.1.1.2",
					NetworkInterface: "fake",
					NodeName:         "fake",
					CloudProvider:    "fake",
					Pool:             "fake",
				},
				{
					Address:          "1.1.1.3",
					NetworkInterface: "fake",
					NodeName:         "fake",
					CloudProvider:    "fake",
					Pool:             "fake",
				},
			},
		},
		Status: loadbalancing_v1alpha1.IPAllocationStatus{
			State:   loadbalancing_v1alpha1.AllocationStatusSuccess,
			Message: "Allocated",
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	type args struct {
		allocationRO   *loadbalancing_v1alpha1.IPAllocation
		removedAddress string
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should remove",
			args: args{
				allocationRO:   &o,
				removedAddress: "1.1.1.1",
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Type: loadbalancing_v1alpha1.PersistentIP,
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.2",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
						{
							Address:          "1.1.1.3",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
					},
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusAddrDeleted,
					Message: fmt.Sprintf("Address %s removed from pool", "1.1.1.1"),
				},
			},
		},
		{
			name: "should remove",
			args: args{
				allocationRO:   &o,
				removedAddress: "1.1.1.2",
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Type: loadbalancing_v1alpha1.PersistentIP,
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.1",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
						{
							Address:          "1.1.1.3",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
					},
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusAddrDeleted,
					Message: fmt.Sprintf("Address %s removed from pool", "1.1.1.2"),
				},
			},
		},
		{
			name: "should remove",
			args: args{
				allocationRO:   &o,
				removedAddress: "1.1.1.3",
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Spec: loadbalancing_v1alpha1.IPAllocationSpec{
					Type: loadbalancing_v1alpha1.PersistentIP,
					Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
						{
							Address:          "1.1.1.1",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
						{
							Address:          "1.1.1.2",
							NetworkInterface: "fake",
							NodeName:         "fake",
							CloudProvider:    "fake",
							Pool:             "fake",
						},
					},
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusAddrDeleted,
					Message: fmt.Sprintf("Address %s removed from pool", "1.1.1.3"),
				},
			},
		},
		{
			name: "should not remove",
			args: args{
				allocationRO:   &o,
				removedAddress: "1.1.1.5",
			},
			want: &o,
		},
		{
			name: "should fail",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_wrong",
						Namespace: "test_namespace",
					},
					Spec: loadbalancing_v1alpha1.IPAllocationSpec{
						Type: loadbalancing_v1alpha1.PersistentIP,
						Allocations: []*loadbalancing_v1alpha1.IPAllocationAddresses{
							{
								Address:          "1.1.1.1",
								NetworkInterface: "fake",
								NodeName:         "fake",
								CloudProvider:    "fake",
								Pool:             "fake",
							},
						},
					},
				},
				removedAddress: "1.1.1.1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RemoveAddressFromAllocation(tt.args.allocationRO, tt.args.removedAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveAddressFromAllocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveAddressFromAllocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteAllocationByName(t *testing.T) {
	allocationNamespace := "test_namespace"
	allocationName := "test_name"
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      allocationName,
			Namespace: allocationNamespace,
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	type args struct {
		allocationNamespace string
		allocationName      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should delete",
			args: args{
				allocationNamespace: allocationNamespace,
				allocationName:      allocationName,
			},
		},
		{
			name: "should fail",
			args: args{
				allocationNamespace: "wrong",
				allocationName:      allocationName,
			},
		},
		{
			name: "should fail",
			args: args{
				allocationNamespace: allocationNamespace,
				allocationName:      "wrong",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteAllocationByName(tt.args.allocationNamespace, tt.args.allocationName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllocationByName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindAllocation(t *testing.T) {

	allocationNamespace := "test_namespace"
	allocationName := "test_name"
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      allocationName,
			Namespace: allocationNamespace,
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	type args struct {
		allocationNamespace string
		allocationName      string
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should find",
			args: args{
				allocationNamespace: allocationNamespace,
				allocationName:      allocationName,
			},
			want: &o,
		},
		{
			name: "should not find",
			args: args{
				allocationNamespace: "wrong",
				allocationName:      allocationName,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAllocation(tt.args.allocationNamespace, tt.args.allocationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAllocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAllocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAllocationStatusFailed(t *testing.T) {
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test_name",
			Namespace: "test_namespace",
		},
	}

	mockGetPlenuslbClient(o.DeepCopyObject())

	reason := errors.New("Reason for the error")
	type args struct {
		allocationRO *loadbalancing_v1alpha1.IPAllocation
		reason       error
	}
	tests := []struct {
		name    string
		args    args
		want    *loadbalancing_v1alpha1.IPAllocation
		wantErr bool
	}{
		{
			name: "should update",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name",
						Namespace: "test_namespace",
					},
				},
				reason: reason,
			},
			want: &loadbalancing_v1alpha1.IPAllocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
				Status: loadbalancing_v1alpha1.IPAllocationStatus{
					State:   loadbalancing_v1alpha1.AllocationStatusFailed,
					Message: reason.Error(),
				},
			},
		},
		{
			name: "should fail with not found error",
			args: args{
				allocationRO: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test_name_wrong",
						Namespace: "test_namespace",
					},
				},
				reason: reason,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetAllocationStatusFailed(tt.args.allocationRO, tt.args.reason)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAllocationStatusFailed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetAllocationStatusFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

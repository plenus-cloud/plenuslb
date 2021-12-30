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

// https://github.com/kubernetes/kubernetes/issues/54075#issuecomment-337298950

package allocationswatcher

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	plenuslbclientset "plenus.io/plenuslb/pkg/client/clientset/versioned"
	plenuslbclientsetfake "plenus.io/plenuslb/pkg/client/clientset/versioned/fake"
	"plenus.io/plenuslb/pkg/controller/clients"
)

var originalGetPlenuslbClient = clients.GetPlenuslbClient

func mockGetPlenuslbClient(objects ...runtime.Object) {
	clients.GetPlenuslbClient = func() plenuslbclientset.Interface {
		return plenuslbclientsetfake.NewSimpleClientset(objects...)
	}
}

func restoreGetPlenuslbClient() {
	clients.GetPlenuslbClient = originalGetPlenuslbClient
}

func TestFindAllocationByNameApi(t *testing.T) {
	allocationNamespace := "test_namespace"
	allocationName := "test_name"
	o := loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      allocationName,
			Namespace: allocationNamespace,
		},
	}
	mockGetPlenuslbClient(o.DeepCopyObject())
	defer restoreGetPlenuslbClient()

	type args struct {
		namespace string
		name      string
	}
	tests := []struct {
		name      string
		args      args
		want      *loadbalancing_v1alpha1.IPAllocation
		wantErr   bool
		errorType error
	}{
		{
			name: "should find using api",
			args: args{
				namespace: allocationNamespace,
				name:      allocationName,
			},
			want: &o,
		},
		{
			name: "should not find using api",
			args: args{
				namespace: allocationNamespace,
				name:      "wrong",
			},
			wantErr:   true,
			errorType: ErrAllocationNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAllocationByName(tt.args.namespace, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAllocationByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.errorType != nil {
				if err != tt.errorType {
					t.Errorf("FindAllocationByName() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAllocationByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

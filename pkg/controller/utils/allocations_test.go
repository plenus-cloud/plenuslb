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

package utils

import (
	"reflect"
	"testing"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

func TestContainsAddress(t *testing.T) {
	type args struct {
		a []*loadbalancing_v1alpha1.IPAllocationAddresses
		x string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 *loadbalancing_v1alpha1.IPAllocationAddresses
	}{
		{
			name: "should not find",
			args: args{
				a: []*loadbalancing_v1alpha1.IPAllocationAddresses{{}},
				x: "1.1.1.1",
			},
			want: false,
		},
		{
			name: "should not find",
			args: args{
				a: []*loadbalancing_v1alpha1.IPAllocationAddresses{
					{Address: "1.2.3.4"},
					{Address: "1.2.5.4"},
				},
				x: "1.1.1.1",
			},
			want: false,
		},
		{
			name: "should find",
			args: args{
				a: []*loadbalancing_v1alpha1.IPAllocationAddresses{
					{Address: "1.2.3.4"},
					{Address: "1.1.1.1", Pool: "Something"},
				},
				x: "1.1.1.1",
			},
			want: true,
			want1: &loadbalancing_v1alpha1.IPAllocationAddresses{
				Address: "1.1.1.1",
				Pool:    "Something",
			},
		},
		{
			name: "should find",
			args: args{
				a: []*loadbalancing_v1alpha1.IPAllocationAddresses{
					{Address: "1.1.1.1", Pool: "Something"},
				},
				x: "1.1.1.1",
			},
			want: true,
			want1: &loadbalancing_v1alpha1.IPAllocationAddresses{
				Address: "1.1.1.1",
				Pool:    "Something",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ContainsAddress(tt.args.a, tt.args.x)
			if got != tt.want {
				t.Errorf("ContainsAddress() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ContainsAddress() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

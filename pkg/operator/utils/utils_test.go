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
	"testing"

	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
)

func TestContainsAddressInfo(t *testing.T) {
	type args struct {
		a             []*plenuslbV1Alpha1.AddressInfo
		interfaceName string
		address       string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should contain",
			args: args{
				a: []*plenuslbV1Alpha1.AddressInfo{
					&plenuslbV1Alpha1.AddressInfo{
						Address:   "10.10.10.20",
						Interface: "interface",
					},
					&plenuslbV1Alpha1.AddressInfo{
						Address:   "10.10.10.30",
						Interface: "interface",
					},
				},
				interfaceName: "interface",
				address:       "10.10.10.30",
			},
			want: true,
		},
		{
			name: "should contain",
			args: args{
				a: []*plenuslbV1Alpha1.AddressInfo{
					&plenuslbV1Alpha1.AddressInfo{
						Address:   "10.10.10.20",
						Interface: "interface",
					},
					&plenuslbV1Alpha1.AddressInfo{
						Address:   "10.10.10.30",
						Interface: "interface2",
					},
				},
				interfaceName: "interface",
				address:       "10.10.10.20",
			},
			want: true,
		},
		{
			name: "should NOT contain",
			args: args{
				a: []*plenuslbV1Alpha1.AddressInfo{
					&plenuslbV1Alpha1.AddressInfo{
						Address:   "10.10.10.20",
						Interface: "interface",
					},
					&plenuslbV1Alpha1.AddressInfo{
						Address:   "10.10.10.30",
						Interface: "interface2",
					},
				},
				interfaceName: "interface",
				address:       "10.10.10.30",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsAddressInfo(tt.args.a, tt.args.interfaceName, tt.args.address); got != tt.want {
				t.Errorf("ContainsAddressInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

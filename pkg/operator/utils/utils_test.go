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

package utils

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

func TestServiceIsLoadBalancer(t *testing.T) {
	type args struct {
		service *v1.Service
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should not be a LoadBalancer",
			args: args{
				service: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeClusterIP,
					},
				},
			},
			want: false,
		},
		{
			name: "should be a LoadBalancer",
			args: args{
				service: &v1.Service{
					Spec: v1.ServiceSpec{
						Type: v1.ServiceTypeLoadBalancer,
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ServiceIsLoadBalancer(tt.args.service); got != tt.want {
				t.Errorf("ServiceIsLoadBalancer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceHasExternalIPs(t *testing.T) {
	type args struct {
		service *v1.Service
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 []string
	}{
		{
			name: "should not have external ips",
			args: args{
				service: &v1.Service{},
			},
			want:  false,
			want1: []string{},
		},
		{
			name: "should not have external ips",
			args: args{
				service: &v1.Service{
					Spec: v1.ServiceSpec{
						ExternalIPs: []string{},
					},
				},
			},
			want:  false,
			want1: []string{},
		},
		{
			name: "should have external ips",
			args: args{
				service: &v1.Service{
					Spec: v1.ServiceSpec{
						ExternalIPs: []string{"1.1.1.1"},
					},
				},
			},
			want:  true,
			want1: []string{"1.1.1.1"},
		},
		{
			name: "should have external ips",
			args: args{
				service: &v1.Service{
					Spec: v1.ServiceSpec{
						ExternalIPs: []string{"1.1.1.1", "2.2.2.2"},
					},
				},
			},
			want:  true,
			want1: []string{"1.1.1.1", "2.2.2.2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ServiceHasExternalIPs(tt.args.service)
			if got != tt.want {
				t.Errorf("ServiceHasExternalIPs() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ServiceHasExternalIPs() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPoolHasAddress(t *testing.T) {
	type args struct {
		pool    *loadbalancing_v1alpha1.PersistentIPPool
		address string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should not contain address",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{
					Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
						Addresses: []string{},
					},
				},
				address: "1.1.1.1",
			},
			want: false,
		},
		{
			name: "should not coutain address",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{
					Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
						Addresses: []string{"2.2.2.2", "3.3.3.3"},
					},
				},
				address: "1.1.1.1",
			},
			want: false,
		},
		{
			name: "should countain address",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{
					Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
						Addresses: []string{"2.2.2.2", "3.3.3.3"},
					},
				},
				address: "2.2.2.2",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PoolHasAddress(tt.args.pool, tt.args.address); got != tt.want {
				t.Errorf("PoolHasAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

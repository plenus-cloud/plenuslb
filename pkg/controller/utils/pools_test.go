package utils

import (
	"testing"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

func TestPersistentPoolHasHostNetworkOption(t *testing.T) {
	type args struct {
		pool *loadbalancing_v1alpha1.PersistentIPPool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should not have host network option",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{},
			},
			want: false,
		},
		{
			name: "should not have host network option",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{
					Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
						Options: &loadbalancing_v1alpha1.PoolOptions{
							HostNetworkInterface: &loadbalancing_v1alpha1.HostNetworkInterfaceOptions{},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "should have host network option",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{
					Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
						Options: &loadbalancing_v1alpha1.PoolOptions{
							HostNetworkInterface: &loadbalancing_v1alpha1.HostNetworkInterfaceOptions{
								AddAddressesToInterface: true,
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PersistentPoolHasHostNetworkOption(tt.args.pool); got != tt.want {
				t.Errorf("PersistentPoolHasHostNetworkOption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEphemeralPoolHasHostNetworkOption(t *testing.T) {
	type args struct {
		pool *loadbalancing_v1alpha1.EphemeralIPPool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should not have host network option",
			args: args{
				pool: &loadbalancing_v1alpha1.EphemeralIPPool{},
			},
			want: false,
		},
		{
			name: "should not have host network option",
			args: args{
				pool: &loadbalancing_v1alpha1.EphemeralIPPool{
					Spec: loadbalancing_v1alpha1.EphemeralIPPoolSpec{
						Options: &loadbalancing_v1alpha1.PoolOptions{
							HostNetworkInterface: &loadbalancing_v1alpha1.HostNetworkInterfaceOptions{},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "should have host network option",
			args: args{
				pool: &loadbalancing_v1alpha1.EphemeralIPPool{
					Spec: loadbalancing_v1alpha1.EphemeralIPPoolSpec{
						Options: &loadbalancing_v1alpha1.PoolOptions{
							HostNetworkInterface: &loadbalancing_v1alpha1.HostNetworkInterfaceOptions{
								AddAddressesToInterface: true,
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EphemeralPoolHasHostNetworkOption(tt.args.pool); got != tt.want {
				t.Errorf("EphemeralPoolHasHostNetworkOption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPersistentPoolHasCloudIntegrationOption(t *testing.T) {
	type args struct {
		pool *loadbalancing_v1alpha1.PersistentIPPool
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			name: "should not have cloud integration option",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{},
			},
			want:  false,
			want1: "",
		},
		{
			name: "should not have cloud integration option",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{
					Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
						CloudIntegration: &loadbalancing_v1alpha1.CloudIntegrations{},
					},
				},
			},
			want:  false,
			want1: "",
		},
		{
			name: "should have cloud integration option",
			args: args{
				pool: &loadbalancing_v1alpha1.PersistentIPPool{
					Spec: loadbalancing_v1alpha1.PersistentIPPoolSpec{
						CloudIntegration: &loadbalancing_v1alpha1.CloudIntegrations{
							Hetzner: &loadbalancing_v1alpha1.HetznerCloud{},
						},
					},
				},
			},
			want:  true,
			want1: "hetzner",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := PersistentPoolHasCloudIntegrationOption(tt.args.pool)
			if got != tt.want {
				t.Errorf("PersistentPoolHasCloudIntegrationOption() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("PersistentPoolHasCloudIntegrationOption() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestEphemeralPoolHasCloudIntegrationOption(t *testing.T) {
	type args struct {
		pool *loadbalancing_v1alpha1.EphemeralIPPool
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			name: "should not have cloud integration option",
			args: args{
				pool: &loadbalancing_v1alpha1.EphemeralIPPool{},
			},
			want:  false,
			want1: "",
		},
		{
			name: "should not have cloud integration option",
			args: args{
				pool: &loadbalancing_v1alpha1.EphemeralIPPool{
					Spec: loadbalancing_v1alpha1.EphemeralIPPoolSpec{
						CloudIntegration: &loadbalancing_v1alpha1.CloudIntegrations{},
					},
				},
			},
			want:  false,
			want1: "",
		},
		{
			name: "should have cloud integration option",
			args: args{
				pool: &loadbalancing_v1alpha1.EphemeralIPPool{
					Spec: loadbalancing_v1alpha1.EphemeralIPPoolSpec{
						CloudIntegration: &loadbalancing_v1alpha1.CloudIntegrations{
							Hetzner: &loadbalancing_v1alpha1.HetznerCloud{},
						},
					},
				},
			},
			want:  true,
			want1: "hetzner",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := EphemeralPoolHasCloudIntegrationOption(tt.args.pool)
			if got != tt.want {
				t.Errorf("EphemeralPoolHasCloudIntegrationOption() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("EphemeralPoolHasCloudIntegrationOption() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

package ephemeralips

import (
	"fmt"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	fakeclietset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	plenuslbclientset "plenus.io/plenuslb/pkg/client/clientset/versioned"
	plenuslbclientsetfake "plenus.io/plenuslb/pkg/client/clientset/versioned/fake"
	"plenus.io/plenuslb/pkg/clouds/fake"
	"plenus.io/plenuslb/pkg/controller/clients"
	"plenus.io/plenuslb/pkg/controller/operator"
)

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

func mockEphemeralPoolCache(pools ...*loadbalancing_v1alpha1.EphemeralIPPool) {
	testStoreKeyFunc := func(obj interface{}) (string, error) {
		return obj.(*loadbalancing_v1alpha1.EphemeralIPPool).Name, nil
	}
	newStore := cache.NewStore(testStoreKeyFunc)

	for _, pool := range pools {
		newStore.Add(pool)
	}
	poolStoreList = func() []interface{} {
		return newStore.List()
	}
}

func mockCoreClientCache(pods ...*v1.Pod) {
	testStoreKeyFunc := func(obj interface{}) (string, error) {
		return fmt.Sprintf("%s/%s", obj.(*v1.Pod).Name, obj.(*v1.Pod).Namespace), nil
	}
	newStore := cache.NewStore(testStoreKeyFunc)

	for _, pod := range pods {
		newStore.Add(pod)
	}
	operator.GetStoreList = func() []interface{} {
		return newStore.List()
	}
}
func mockCloudsIntegration() {
	cloudsIntegration = &fake.Integration{}
}

func Test_buildAllocations(t *testing.T) {
	mockCloudsIntegration()
	node := v1.Node{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "fakeNodeName",
		},
	}
	mockGetK8sClient(node.DeepCopyObject())

	poolWithCloud := loadbalancing_v1alpha1.EphemeralIPPool{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test_ephemeral",
		},
		Spec: loadbalancing_v1alpha1.EphemeralIPPoolSpec{
			AllowedNamespaces: []string{"with-cloud-with-host"},
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

	poolWithoutCloud := loadbalancing_v1alpha1.EphemeralIPPool{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test_ephemeral_no_cloud",
		},
		Spec: loadbalancing_v1alpha1.EphemeralIPPoolSpec{
			AllowedNamespaces: []string{"without-cloud"},
			Options: &loadbalancing_v1alpha1.PoolOptions{
				HostNetworkInterface: &loadbalancing_v1alpha1.HostNetworkInterfaceOptions{
					AddAddressesToInterface: true,
					InterfaceName:           "test",
				},
			},
		},
	}

	poolWithoutHost := loadbalancing_v1alpha1.EphemeralIPPool{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test_ephemeral_no_host",
		},
		Spec: loadbalancing_v1alpha1.EphemeralIPPoolSpec{
			AllowedNamespaces: []string{"without-host"},
			CloudIntegration: &loadbalancing_v1alpha1.CloudIntegrations{
				Hetzner: &loadbalancing_v1alpha1.HetznerCloud{
					Token: "fake_token",
				},
			},
		},
	}

	poolWithoutCloudAndHost := loadbalancing_v1alpha1.EphemeralIPPool{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "test_ephemeral_no_cloud_no_host",
		},
		Spec: loadbalancing_v1alpha1.EphemeralIPPoolSpec{
			AllowedNamespaces: []string{"no-cloud-no-host"},
		},
	}

	mockEphemeralPoolCache(&poolWithoutCloudAndHost, &poolWithoutHost, &poolWithoutCloud, &poolWithCloud)

	operatorPod := v1.Pod{
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

	mockCoreClientCache(&operatorPod)
	type args struct {
		serviceNamespace string
		serviceName      string
	}
	tests := []struct {
		name    string
		args    args
		want    []*loadbalancing_v1alpha1.IPAllocationAddresses
		wantErr bool
	}{
		{
			name: "Should fail due pool not found",
			args: args{
				serviceName:      "fake_name",
				serviceNamespace: "fake_namespace",
			},
			want:    []*loadbalancing_v1alpha1.IPAllocationAddresses{},
			wantErr: true,
		},
		{
			name: "Should succeed with cloud and host",
			args: args{
				serviceName:      "fake_name",
				serviceNamespace: "with-cloud-with-host",
			},
			want: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address:          "1.1.1.1",
					Pool:             "test_ephemeral",
					NetworkInterface: "test",
					CloudProvider:    "hetzner",
					NodeName:         "fakePodNodeName",
				},
			},
			wantErr: false,
		},
		{
			name: "Should succeed with cloud",
			args: args{
				serviceName:      "fake_name",
				serviceNamespace: "without-host",
			},
			want: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address:       "1.1.1.1",
					Pool:          "test_ephemeral_no_host",
					CloudProvider: "hetzner",
					NodeName:      "fakeNodeName",
				},
			},
			wantErr: false,
		},
		{
			name: "Should succeed with host",
			args: args{
				serviceName:      "fake_name",
				serviceNamespace: "without-cloud",
			},
			want: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address:          "",
					Pool:             "test_ephemeral_no_cloud",
					NetworkInterface: "test",
					NodeName:         "fakePodNodeName",
				},
			},
			wantErr: false,
		},
		{
			name: "Should succeed no host no cloud",
			args: args{
				serviceName:      "fake_name",
				serviceNamespace: "no-cloud-no-host",
			},
			want: []*loadbalancing_v1alpha1.IPAllocationAddresses{
				{
					Address: "",
					Pool:    "test_ephemeral_no_cloud_no_host",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAllocations(tt.args.serviceNamespace, tt.args.serviceName)
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

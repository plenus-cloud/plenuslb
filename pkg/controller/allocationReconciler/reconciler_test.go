package allocationreconciler

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	fakeclietset "k8s.io/client-go/kubernetes/fake"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	plenuslbclientset "plenus.io/plenuslb/pkg/client/clientset/versioned"
	plenuslbclientsetfake "plenus.io/plenuslb/pkg/client/clientset/versioned/fake"
	"plenus.io/plenuslb/pkg/controller/clients"
)

func mockGetK8sClient(objects ...runtime.Object) {
	clients.GetK8sClient = func() clientset.Interface {
		return fakeclietset.NewSimpleClientset(objects...)
	}
}

func mockGetPlenuslbClient(objects ...runtime.Object) {
	clients.GetPlenuslbClient = func() plenuslbclientset.Interface {
		return plenuslbclientsetfake.NewSimpleClientset(objects...)
	}
}

var (
	serviceName      = "service-name"
	serviceNamespace = "service-namespace"
	serviceMock      = &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
		},
	}

	ephemeralServiceMock = &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      serviceName + "-ephemeral",
			Namespace: serviceNamespace,
		},
	}

	allocationMock = &loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
		},
	}

	ephemeralAllocationMock = &loadbalancing_v1alpha1.IPAllocation{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      serviceName + "-ephemeral",
			Namespace: serviceNamespace,
		},
		Spec: loadbalancing_v1alpha1.IPAllocationSpec{
			Type: loadbalancing_v1alpha1.EphemeralIP,
		},
	}
)

func Test_expectedAllocationType(t *testing.T) {
	type args struct {
		service *v1.Service
	}
	tests := []struct {
		name string
		args args
		want loadbalancing_v1alpha1.IPType
	}{
		{
			name: "should return persistent ip",
			args: args{
				service: &v1.Service{
					Spec: v1.ServiceSpec{
						ExternalIPs: []string{
							"1.1.1.1",
						},
					},
				},
			},
			want: loadbalancing_v1alpha1.PersistentIP,
		},
		{
			name: "should return ephemeral ip",
			args: args{
				service: &v1.Service{},
			},
			want: loadbalancing_v1alpha1.EphemeralIP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expectedAllocationType(tt.args.service); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expectedAllocationType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serviceIsNoLongerALoadBalancer(t *testing.T) {
	mockGetK8sClient(ephemeralServiceMock.DeepCopyObject(), serviceMock.DeepCopyObject())
	mockGetPlenuslbClient(allocationMock.DeepCopyObject(), ephemeralAllocationMock.DeepCopyObject())

	type args struct {
		service    *v1.Service
		allocation *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should delete allocation",
			args: args{
				service:    serviceMock,
				allocation: allocationMock,
			},
		},
		{
			name: "should delete allocation",
			args: args{
				service:    ephemeralServiceMock,
				allocation: ephemeralAllocationMock,
			},
		},
		// TODO: what is going on here?
		/* {
			name: "should fail due service not found",
			args: args{
				service: &v1.Service{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "wrong",
						Namespace: serviceNamespace,
					},
				},
				allocation: allocationMock,
			},
			wantErr: true,
		},
		{
			name: "should fail due allocation not found",
			args: args{
				service: serviceMock,
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "wrong",
						Namespace: serviceNamespace,
					},
				},
			},
			wantErr: true,
		}, */
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := serviceIsNoLongerALoadBalancer(tt.args.service, tt.args.allocation); (err != nil) != tt.wantErr {
				t.Errorf("serviceIsNoLongerALoadBalancer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

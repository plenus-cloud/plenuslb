package allocationslock

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/utils"
)

func mockDefaultObjectBackoff() {
	utils.DefaultObjectBackoff = wait.Backoff{
		Steps:    1,
		Duration: 1 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
	}
}

func resetLock() {
	lockedObjects = sync.Map{}
}

func TestAcquireAllocationLock(t *testing.T) {
	mockDefaultObjectBackoff()
	type args struct {
		allocation *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		before  func()
	}{
		{
			name: "should acquire",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "name",
						Namespace: "namespace",
					},
				},
			},
		},
		{
			name: "should not acquire",
			before: func() {
				lockedObjects.Store(fmt.Sprintf("%s/%s", "namespace", "name"), "TestAcquireAllocationLock")
			},
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "name",
						Namespace: "namespace",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				tt.before()
			}
			if err := AcquireAllocationLock(tt.args.allocation); (err != nil) != tt.wantErr {
				t.Errorf("AcquireAllocationLock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveFromLock(t *testing.T) {
	type args struct {
		allocation *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name               string
		args               args
		before             func()
		test               func() bool
		expectedTestResult bool
	}{
		{
			name: "should remove",
			before: func() {
				lockedObjects.Store(fmt.Sprintf("%s/%s", "namespace", "name"), nil)
			},
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "name",
						Namespace: "namespace",
					},
				},
			},
			test: func() bool {
				_, ok := lockedObjects.Load(fmt.Sprintf("%s/%s", "namespace", "name"))
				return ok
			},
			expectedTestResult: false,
		},
		{
			name: "should remove",
			before: func() {
				lockedObjects.Store(fmt.Sprintf("%s/%s", "namespace", "name"), nil)
				lockedObjects.Store(fmt.Sprintf("%s/%s", "namespace", "name2"), nil)
			},
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "name2",
						Namespace: "namespace",
					},
				},
			},
			test: func() bool {
				_, ok := lockedObjects.Load(fmt.Sprintf("%s/%s", "namespace", "name"))
				return ok
			},
			expectedTestResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				tt.before()
			}
			RemoveFromLock(tt.args.allocation)
			if testRes := tt.test(); tt.expectedTestResult != testRes {
				t.Errorf("RemoveFromLock() tt.test() = %v, expectedTestResult %v", testRes, tt.expectedTestResult)
			}
		})
	}
}

func TestIsErrorAllocationAlreadyProcessing(t *testing.T) {
	errorAllocationsBackoffDict = []string{"test-ns/test", "test-ns/test2"}
	type args struct {
		allocation *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is processing",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
				},
			},
			want: true,
		},
		{
			name: "is processing",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test2",
						Namespace: "test-ns",
					},
				},
			},
			want: true,
		},
		{
			name: "is not processing",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test3",
						Namespace: "test-ns",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsErrorAllocationAlreadyProcessing(tt.args.allocation); got != tt.want {
				t.Errorf("IsErrorAllocationAlreadyProcessing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddErrorAllocationToProcessingList(t *testing.T) {
	errorAllocationsBackoffDict = []string{"test-ns/test"}
	type args struct {
		allocation *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should add",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test2",
						Namespace: "test-ns",
					},
				},
			},
			want: []string{"test-ns/test", "test-ns/test2"},
		},
		{
			name: "should not add",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
				},
			},
			want: []string{"test-ns/test", "test-ns/test2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddErrorAllocationToProcessingList(tt.args.allocation); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddErrorAllocationToProcessingList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveErrorAllocationFromProcessingList(t *testing.T) {
	errorAllocationsBackoffDict = []string{"test-ns/test", "test-ns/test2", "test-ns/test3"}
	type args struct {
		allocation *loadbalancing_v1alpha1.IPAllocation
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should not remove",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test4",
						Namespace: "test-ns",
					},
				},
			},
			want: []string{"test-ns/test", "test-ns/test2", "test-ns/test3"},
		},
		{
			name: "should remove",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test2",
						Namespace: "test-ns",
					},
				},
			},
			want: []string{"test-ns/test", "test-ns/test3"},
		},
		{
			name: "should remove",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test",
						Namespace: "test-ns",
					},
				},
			},
			want: []string{"test-ns/test3"},
		},
		{
			name: "should remove",
			args: args{
				allocation: &loadbalancing_v1alpha1.IPAllocation{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "test3",
						Namespace: "test-ns",
					},
				},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := RemoveErrorAllocationFromProcessingList(tt.args.allocation); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddErrorAllocationToProcessingList() = %v, want %v", got, tt.want)
			}
		})
	}
}

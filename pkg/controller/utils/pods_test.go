package utils

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestIsPodReady(t *testing.T) {
	type args struct {
		pod *v1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should not be ready",
			args: args{
				pod: &v1.Pod{},
			},
			want: false,
		},
		{
			name: "should not be ready",
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type: v1.PodScheduled,
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "should not be ready",
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type: v1.PodReady,
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "should be ready",
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   v1.PodReady,
								Status: v1.ConditionTrue,
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "should be ready",
			args: args{
				pod: &v1.Pod{
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   v1.PodReady,
								Status: v1.ConditionTrue,
							},
							{
								Type:   v1.PodScheduled,
								Status: v1.ConditionFalse,
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
			if got := IsPodReady(tt.args.pod); got != tt.want {
				t.Errorf("IsPodReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

package utils

import v1 "k8s.io/api/core/v1"

// IsPodReady returns true if the given pod is up and running
func IsPodReady(pod *v1.Pod) bool {
	if &pod.Status != nil && len(pod.Status.Conditions) > 0 {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady &&
				condition.Status == v1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

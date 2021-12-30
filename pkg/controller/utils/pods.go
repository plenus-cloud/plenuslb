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

import v1 "k8s.io/api/core/v1"

// IsPodReady returns true if the given pod is up and running
func IsPodReady(pod *v1.Pod) bool {
	if len(pod.Status.Conditions) > 0 {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady &&
				condition.Status == v1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

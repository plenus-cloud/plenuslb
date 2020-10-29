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
	"os"
	"time"

	"github.com/Pallinder/go-randomdata"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
)

// ServiceIsLoadBalancer returns true if the given service is of type load balancer
func ServiceIsLoadBalancer(service *v1.Service) bool {
	if service != nil && service.Spec.Type == v1.ServiceTypeLoadBalancer {
		return true
	}
	return false
}

// ServiceHasExternalIPs return true if the service has at least one external ip
func ServiceHasExternalIPs(service *v1.Service) (bool, []string) {
	if len(service.Spec.ExternalIPs) > 0 {
		return true, service.Spec.ExternalIPs
	}
	return false, []string{}
}

// ContainsString tells whether a contains x.
func ContainsString(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// PoolHasAddress checks if the given persistent ip pool ad a specific address
func PoolHasAddress(pool *loadbalancing_v1alpha1.PersistentIPPool, address string) bool {
	for _, poolAddress := range pool.Spec.Addresses {
		if poolAddress == address {
			return true
		}
	}
	return false
}

// GetClusterName returns a cluster name from the CLUSTER_NAME
// or a silly name if the env variable is not provided
func GetClusterName() string {
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		return randomdata.SillyName()
	}
	return clusterName
}

// DefaultObjectBackoff is the default backoff for a generic object
var DefaultObjectBackoff = wait.Backoff{
	Steps:    10,
	Duration: 100 * time.Millisecond,
	Factor:   2.0,
	Jitter:   0.1,
}

// DefaultAPIBackoff is the default backoff for external apis
var DefaultAPIBackoff = wait.Backoff{
	Steps:    10,
	Duration: 10 * time.Millisecond,
	Factor:   2.0,
	Jitter:   0.1,
}

// ForeverBackoff holds parameters applied to a Backoff function.
type ForeverBackoff struct {
	// The initial duration.
	Duration time.Duration
	// Duration is multiplied by factor each iteration, if factor is not zero
	// and the limits imposed by Steps and Cap have not been reached.
	// Should not be negative.
	// The jitter does not contribute to the updates to the duration parameter.
	Factor float64
	// The sleep at each iteration is the duration plus an additional
	// amount chosen uniformly at random from the interval between
	// zero and `jitter*duration`.
	Jitter float64
	// The remaining number of iterations in which the duration
	// parameter may change (but progress can be stopped earlier by
	// hitting the cap). If not positive, the duration is not
	// changed. Used for exponential backoff in combination with
	// Factor and Cap.
	Steps int
	// A limit on revised values of the duration parameter. If a
	// multiplication by the factor parameter would make the duration
	// exceed the cap then the duration is set to the cap
	Cap time.Duration
}

// Step (1) returns an amount of time to sleep determined by the
// original Duration and Jitter and (2) mutates the provided Backoff
// to update its Steps and Duration.
func (b *ForeverBackoff) Step() time.Duration {
	if b.Steps < 1 {
		if b.Jitter > 0 {
			return wait.Jitter(b.Duration, b.Jitter)
		}
		return b.Duration
	}
	b.Steps--

	duration := b.Duration

	// calculate the next step
	if b.Factor != 0 {
		b.Duration = time.Duration(float64(b.Duration) * b.Factor)
		if b.Cap > 0 && b.Duration > b.Cap {
			b.Duration = b.Cap
		}
	}

	if b.Jitter > 0 {
		duration = wait.Jitter(duration, b.Jitter)
	}
	return duration
}

// ExponentialBackoffWithForeverCap repeats a condition check with exponential backoff.
//
// It repeatedly checks the condition and then sleeps, using `backoff.Step()`
// to determine the length of the sleep and adjust Duration and Steps.
// Stops and returns as soon as:
// 1. the condition check returns true or an error,
// 2. `backoff.Steps` checks of the condition have been done, or
// 3. a sleep truncated by the cap on duration has been completed.
// In case (1) the returned error is what the condition function returned.
// In all other cases, ErrWaitTimeout is returned.
func ExponentialBackoffWithForeverCap(backoff ForeverBackoff, condition wait.ConditionFunc) error {
	for backoff.Steps > 0 {
		if ok, err := condition(); err != nil || ok {
			return err
		}
		if backoff.Steps == 1 {
			break
		}
		step := backoff.Step()
		time.Sleep(step)
	}
	return wait.ErrWaitTimeout

}

// OnErrorForever executes the provided function repeatedly, retrying if the server returns a specified
// error. Callers should preserve previous executions if they wish to retry changes. It performs an
// exponential backoff with forever cap.
//
//     var pod *api.Pod
//     err := retry.OnErrorForever(DefaultBackoff, errors.IsConflict, func() (err error) {
//       pod, err = c.Pods("mynamespace").UpdateStatus(podStatus)
//       return
//     })
//     if err != nil {
//       // may be conflict if max retries were hit
//       return err
//     }
//     ...
//
func OnErrorForever(backoff ForeverBackoff, errorFunc func(error) bool, fn func() error) error {
	var lastConflictErr error
	err := ExponentialBackoffWithForeverCap(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		case errorFunc(err):
			lastConflictErr = err
			return false, nil
		default:
			return false, err
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastConflictErr
	}
	return err
}

// ErrorBackoff is the backoff for allocations in error state
var ErrorBackoff = ForeverBackoff{
	Steps: 552,
	// Steps:    20,
	Duration: 1 * time.Second,
	Factor:   1.2,
	Jitter:   0.2,
	Cap:      5 * time.Minute,
}

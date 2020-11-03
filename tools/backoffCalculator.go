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

package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

var maxRetrytime float64

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
		//time.Sleep(backoff.Step())
		maxRetrytime += step.Seconds()
		fmt.Println(step)
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

func plural(count int, singular string) (result string) {
	if (count == 1) || (count == 0) {
		result = strconv.Itoa(count) + " " + singular + " "
	} else {
		result = strconv.Itoa(count) + " " + singular + "s "
	}
	return
}

func secondsToHuman(input int) (result string) {
	years := math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30 / 12)
	seconds := input % (60 * 60 * 24 * 7 * 30 * 12)
	months := math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30)
	seconds = input % (60 * 60 * 24 * 7 * 30)
	weeks := math.Floor(float64(seconds) / 60 / 60 / 24 / 7)
	seconds = input % (60 * 60 * 24 * 7)
	days := math.Floor(float64(seconds) / 60 / 60 / 24)
	seconds = input % (60 * 60 * 24)
	hours := math.Floor(float64(seconds) / 60 / 60)
	seconds = input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = input % 60

	if years > 0 {
		result = plural(int(years), "year") + plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if months > 0 {
		result = plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if weeks > 0 {
		result = plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if days > 0 {
		result = plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if hours > 0 {
		result = plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if minutes > 0 {
		result = plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else {
		result = plural(int(seconds), "second")
	}

	return
}

func floatToInt(f float64) int {
	s := fmt.Sprintf("%.0f", f)
	i, err := strconv.Atoi(s)
	if err == nil {
		return i
	}
	fmt.Println(err)
	return 0
}

func main() {
	var ErrorBackoff = ForeverBackoff{
		//Steps:    552,
		Steps:    20,
		Duration: 1 * time.Second,
		Factor:   1.2,
		Jitter:   0.2,
		Cap:      5 * time.Minute,
	}

	OnErrorForever(ErrorBackoff, func(err error) bool { return true }, func() error { return errors.New("err") })

	fmt.Printf("Total max retry time is %s\n", secondsToHuman(floatToInt(maxRetrytime)))
}

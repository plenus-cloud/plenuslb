package allocationslock

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"k8s.io/client-go/util/retry"
	"k8s.io/klog"

	loadbalancing_v1alpha1 "plenus.io/plenuslb/pkg/apis/loadbalancing/v1alpha1"
	"plenus.io/plenuslb/pkg/controller/utils"
)

// ErrObjectLocked is returned when is tryin to acquire a locked object
var ErrObjectLocked = errors.New("Allocation object is locked, some operation is still in progress")

// lockedObjects saves the objects on which operations are in progress, in order to avoid that more operations are done on the same object at the same time
var lockedObjects sync.Map

func addToLock(allocation *loadbalancing_v1alpha1.IPAllocation, byWho string) {
	klog.Infof("Lock of allocation %s/%s acquired by  %s", allocation.GetNamespace(), allocation.GetName(), byWho)
	lockedObjects.Store(fmt.Sprintf("%s/%s", allocation.GetNamespace(), allocation.GetName()), byWho)
}

func isLocked(allocation *loadbalancing_v1alpha1.IPAllocation) (bool, string) {
	byWho, ok := lockedObjects.Load(fmt.Sprintf("%s/%s", allocation.GetNamespace(), allocation.GetName()))
	byWhoString := ""
	if ok {
		byWhoString = byWho.(string)
	}
	return ok, byWhoString
}

// RemoveFromLock removes the given allocation from the lock list
func RemoveFromLock(allocation *loadbalancing_v1alpha1.IPAllocation) {
	lockedObjects.Delete(fmt.Sprintf("%s/%s", allocation.GetNamespace(), allocation.GetName()))
}

func trace(skip int) string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	s := strings.Split(frame.File, "/")
	file := ""
	if len(s) >= 2 {
		file = fmt.Sprintf("%s/%s", s[len(s)-2], s[len(s)-1])
	} else {
		file = s[len(s)-1]
	}
	trace := fmt.Sprintf("%s:%d", file, frame.Line)
	return trace
}

// AcquireAllocationLock tryies to lock the given object with an exponential backoff. The error can only be a timeout
func AcquireAllocationLock(allocation *loadbalancing_v1alpha1.IPAllocation) error {
	calledBy := trace(3)
	err := retry.OnError(utils.DefaultObjectBackoff, errIsObjectLocked, func() (err error) {
		isLocked, byWho := isLocked(allocation)
		if !isLocked {
			addToLock(allocation, calledBy)
			return nil
		}
		klog.Infof("Allocation %s/%s is locked, some other operations are still in progress. Lock was called by %s, now by %s", allocation.GetNamespace(), allocation.GetName(), byWho, calledBy)
		return ErrObjectLocked
	})

	if err != nil {
		// may be conflict if max retries were hit
		klog.Error(err)
		return err
	}

	return nil
}

func errIsObjectLocked(err error) bool {
	return err == ErrObjectLocked
}

var errorAllocationsBackoffDict []string
var errorAllocationsProcessingMux sync.Mutex

// IsErrorAllocationAlreadyProcessing checks if an allocation in managed by someone else
func IsErrorAllocationAlreadyProcessing(allocation *loadbalancing_v1alpha1.IPAllocation) bool {
	errorAllocationsProcessingMux.Lock()
	defer errorAllocationsProcessingMux.Unlock()

	return isErrorAllocationAlreadyProcessing(allocation)
}

func isErrorAllocationAlreadyProcessing(allocation *loadbalancing_v1alpha1.IPAllocation) bool {
	for _, e := range errorAllocationsBackoffDict {
		if e == fmt.Sprintf("%s/%s", allocation.GetNamespace(), allocation.GetName()) {
			return true
		}
	}
	return false
}

// AddErrorAllocationToProcessingList adds the failed allocation to the "processing" list
func AddErrorAllocationToProcessingList(allocation *loadbalancing_v1alpha1.IPAllocation) []string {
	errorAllocationsProcessingMux.Lock()
	defer errorAllocationsProcessingMux.Unlock()

	if !isErrorAllocationAlreadyProcessing(allocation) {
		klog.Infof("Adding error allocation %s/%s into processing list", allocation.GetNamespace(), allocation.GetName())
		errorAllocationsBackoffDict = append(errorAllocationsBackoffDict, fmt.Sprintf("%s/%s", allocation.GetNamespace(), allocation.GetName()))
	}
	return errorAllocationsBackoffDict
}

// RemoveErrorAllocationFromProcessingList removes the failed allocation from the "processing" list
func RemoveErrorAllocationFromProcessingList(allocation *loadbalancing_v1alpha1.IPAllocation) []string {
	errorAllocationsProcessingMux.Lock()
	defer errorAllocationsProcessingMux.Unlock()

	for i, e := range errorAllocationsBackoffDict {
		if e == fmt.Sprintf("%s/%s", allocation.GetNamespace(), allocation.GetName()) {
			errorAllocationsBackoffDict[i] = errorAllocationsBackoffDict[len(errorAllocationsBackoffDict)-1] // Copy last element to index i.
			errorAllocationsBackoffDict[len(errorAllocationsBackoffDict)-1] = ""                             // Erase last element (write zero value).
			errorAllocationsBackoffDict = errorAllocationsBackoffDict[:len(errorAllocationsBackoffDict)-1]   // Truncate slice.

			klog.Infof("Removed error allocation %s/%s from processing list", allocation.GetNamespace(), allocation.GetName())
			break
		}
	}

	return errorAllocationsBackoffDict
}

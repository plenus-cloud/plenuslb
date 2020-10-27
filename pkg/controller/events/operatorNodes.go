package events

import "k8s.io/klog"

var operatorNodeLost = make(chan string, 100)
var newOperatorNode = make(chan string, 100)

var onNodeLostCB *func(nodeName string)
var onNewNodeCB *func(nodeName string)

// OperatorNodeLost pushes the node lost name into the operator node lost channel
func OperatorNodeLost(nodeName string) {
	operatorNodeLost <- nodeName
}

// RegisterOnOperatorNodeLostFunc registers a callback to be fired when an operator node is lost
func RegisterOnOperatorNodeLostFunc(cb func(nodeName string)) {
	onNodeLostCB = &cb
}

// NewOperatorNode pushes the new operator node name into the new operator node channel
func NewOperatorNode(nodeName string) {
	newOperatorNode <- nodeName
}

//RegisterOnNewOperatorNodeFunc registers a callback to be fired when an operator node is added
func RegisterOnNewOperatorNodeFunc(cb func(nodeName string)) {
	onNewNodeCB = &cb
}

// ListenOperatorNodeLostChan strarts a listener on the channel of operator node lost
// and fires, ad each event, the callback previously registered for this kind of event
func ListenOperatorNodeLostChan(stopCh chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				klog.Warning("Stopping ListenOperatorNodeLostChan goroutine")
				return
			case lost := <-operatorNodeLost:
				if onNodeLostCB != nil {
					(*onNodeLostCB)(lost)
				} else {
					klog.Error("onNodeLostCB is nil, cannot process operator node lost event")
				}
			}
		}
	}()
}

// ListenNewOperatorNodeChan strarts a listener on the channel of new operator node
// and fires, ad each event, the callback previously registered for this kind of event
func ListenNewOperatorNodeChan(stopCh chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				klog.Warning("Stopping ListenNewOperatorNodeChan goroutine")
				return
			case new := <-newOperatorNode:
				if onNewNodeCB != nil {
					(*onNewNodeCB)(new)
				} else {
					klog.Error("newOperatorNode is nil, cannot process new operator node event")
				}
			}
		}
	}()
}

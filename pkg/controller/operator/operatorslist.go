package operator

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"plenus.io/plenuslb/pkg/controller/utils"
	plenuslbV1Alpha1 "plenus.io/plenuslb/pkg/proto/v1alpha1/generated"
)

// OperatorPort is the port on witch the operator server listens
const OperatorPort = 10000

// ErrOperatorNotReady is returned when there's an errol talking with an operator
var ErrOperatorNotReady = errors.New("Operator not ready")

func init() {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
}

// Operator contains the usefull data to talk with al operator node
type Operator struct {
	Address  string
	Port     string
	NodeName string
}

// SearchOperatorByClusterNodeName returns the informtions af the operator on a specific node (if exists)
func SearchOperatorByClusterNodeName(name string) *Operator {
	for _, obj := range operatorNodesStore.List() {
		node, ok := obj.(*v1.Pod)
		if !ok {
			klog.Errorf("unexpected type %s", reflect.TypeOf(obj))
			return nil
		}
		if node.Spec.NodeName == name {
			n := GetNodeFromOperatorPod(node)
			return &n
		}
	}
	return nil
}

// GetOperatorsList returns the list of all the operators's pod
func GetOperatorsList() []interface{} {
	return operatorNodesStore.List()
}

// GetOperatorByName returns the informtions af the operator with the given name (if exists)
func GetOperatorByName(name string) (Operator, error) {
	klog.Infof("Getting operator %s", name)
	for _, obj := range operatorNodesStore.List() {
		operatorPod, ok := obj.(*v1.Pod)
		if !ok {
			err := fmt.Errorf("unexpected type %s", reflect.TypeOf(obj))
			klog.Error(err)
			return Operator{}, err
		}
		if operatorPod.GetName() == name {
			return GetNodeFromOperatorPod(operatorPod), nil
		}
	}

	return Operator{}, utils.ErrNoOperatorNodeAvailable
}

// GetRandomOperatorNode returns a random ready operator
func GetRandomOperatorNode() (*Operator, error) {
	var podNode *v1.Pod
	errIsNotReady := func(err error) bool {
		return err == ErrOperatorNotReady
	}

	err := retry.OnError(retry.DefaultBackoff, errIsNotReady, func() (err error) {
		storeList := GetStoreList()
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s) // initialize local pseudorandom generator
		i := r.Intn(len(storeList))

		obj := storeList[i] // Truncc.poolste slice.

		node, ok := obj.(*v1.Pod)
		if !ok {
			err := fmt.Errorf("unexpected type %s", reflect.TypeOf(obj))
			klog.Error(err)
			return err
		}
		podNode = node

		if !utils.IsPodReady(node) {
			return ErrOperatorNotReady
		}
		return nil
	})
	if err != nil {
		// may be conflict if max retries were hit
		return nil, err
	}

	opNode := GetNodeFromOperatorPod(podNode)
	return &opNode, nil
}

// GetNodeFromOperatorPod builds the operator's connection info from the pod data
func GetNodeFromOperatorPod(operator *v1.Pod) Operator {
	return Operator{
		Address:  operator.Status.PodIP,
		NodeName: operator.Spec.NodeName,
		Port:     strconv.Itoa(OperatorPort),
	}
}

// GrpcClientForNode creates the grpc client from the operator's connection info
func GrpcClientForNode(node Operator) (*grpc.ClientConn, *plenuslbV1Alpha1.PlenusLbClient, error) {
	opts := []grpc.DialOption{grpc.WithBalancerName(roundrobin.Name)}
	opts = append(opts, grpc.WithInsecure())
	var err error
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Address, node.Port), opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
		return nil, nil, err
	}

	client := plenuslbV1Alpha1.NewPlenusLbClient(conn)

	return conn, &client, err
}

// GetOperatorsCount returns the number of the operators
func GetOperatorsCount() int {
	return len(operatorNodesStore.List())
}

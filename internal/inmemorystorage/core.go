package inmemorystorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kube-rise/internal/broadcast"
	"net/http"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
	"strconv"
)

// Structs

type NodeInMemoryStorage struct {
	nodes                      core.NodeList
	nodeEventChan              chan metav1.WatchEvent
	nodeBroadcaster            *broadcast.BroadcastServer[metav1.WatchEvent]
	nodeUpscalingChan          chan core.Node
	nodeDownscalingChan        chan core.Node
	nodeUpscalingBroadcaster   *broadcast.BroadcastServer[core.Node]
	nodeDownscalingBroadcaster *broadcast.BroadcastServer[core.Node]
}

type PodInMemoryStorage struct {
	pods           core.PodList
	podEventChan   chan metav1.WatchEvent
	podBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
	nextResourceId int
}

type NamespaceInMemoryStorage struct {
	namespaces           core.NamespaceList
	namespaceEventChan   chan metav1.WatchEvent
	namespaceBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
}

// Implementations

func (s *NodeInMemoryStorage) GetNodes() (core.NodeList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.nodes, s.nodeBroadcaster
}

func (s *NodeInMemoryStorage) StoreNodes(nodes core.NodeList, events []metav1.WatchEvent) {
	s.nodes = nodes
	for _, n := range events {
		s.nodeEventChan <- n
	}
}

func (s *NodeInMemoryStorage) PutNode(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	w.Header().Set("Content-Type", "application/json")

	// Get node name as path parameter
	pathParams := mux.Vars(r)
	nodeName := pathParams["nodeName"]

	reqBody, _ := io.ReadAll(r.Body)
	// fmt.Println(string(reqBody))
	var u core.Node
	err := json.Unmarshal(reqBody, &u)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		w.WriteHeader(500)
		return
	}

	indexForReplacement := -1
	for index, node := range s.nodes.Items {
		if node.Name == nodeName {
			indexForReplacement = index
			break
		}
	}
	s.nodes.Items[indexForReplacement] = u

	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		fmt.Printf("Unable to encode response for node update, error is: %v", err)
		return
	}
}

func (s *NodeInMemoryStorage) GetNode(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	w.Header().Set("Content-Type", "application/json")

	// Get node name as path parameter
	pathParams := mux.Vars(r)
	nodeName := pathParams["nodeName"]

	var u core.Node

	for _, node := range s.nodes.Items {
		if node.Name == nodeName {
			u = node
			break
		}
	}

	err := json.NewEncoder(w).Encode(u)
	if err != nil {
		fmt.Printf("Unable to encode response for node get, error is: %v", err)
		return
	}
}

func (s *NodeInMemoryStorage) ScaleNodes(addedMachines []cluster.Machine, amount int) ([]core.Node, error) {
	var changedNodes []core.Node
	// Case downscaling
	if amount < 0 {
		amount = amount * -1
		var nodeIndicesToDelete []int
		for i, node := range s.nodes.Items {
			for _, taint := range node.Spec.Taints {
				if taint.Key == "ToBeDeletedByClusterAutoscaler" {
					nodeIndicesToDelete = append(nodeIndicesToDelete, i)
					break
				}
			}
		}
		if len(nodeIndicesToDelete) != amount {
			return []core.Node{}, errors.New(fmt.Sprintf("Mismatch: found %d desired nodes to delete, got %d tainted nodes", amount, len(nodeIndicesToDelete)))
		}
		for amount > 0 {
			nodeToDelete := s.nodes.Items[nodeIndicesToDelete[amount-1]]
			changedNodes = append(changedNodes, nodeToDelete)
			// Delete node from node slice (https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang)
			s.nodes.Items[nodeIndicesToDelete[amount-1]] = s.nodes.Items[len(s.nodes.Items)-1]
			s.nodes.Items = s.nodes.Items[:len(s.nodes.Items)-1]
			amount = amount - 1
		}
		// TODO: Check whether this works (especially pointer to changedNodes should not be overwritten)
		for i, _ := range changedNodes {
			nodeDeleteEvent := metav1.WatchEvent{Type: "DELETED", Object: runtime.RawExtension{Object: &changedNodes[i]}}
			s.nodeEventChan <- nodeDeleteEvent
			s.nodeDownscalingChan <- changedNodes[i]
		}
	} else if amount > 0 {
		newNodes := make([]core.Node, len(addedMachines))
		for i, machine := range addedMachines {
			cpuQuantity, _ := resource.ParseQuantity(machine.Annotations["cpu"])
			memoryQuantity, _ := resource.ParseQuantity(machine.Annotations["memory"])
			podsQuantity, _ := resource.ParseQuantity(machine.Annotations["pods"])
			newNodes[i] = core.Node{
				TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Node"},
				ObjectMeta: metav1.ObjectMeta{Name: machine.Name + "-node"},
				Spec:       core.NodeSpec{ProviderID: "clusterapi://" + machine.Name},
				Status: core.NodeStatus{Phase: "Running", Conditions: []core.NodeCondition{
					{
						Type:   "Ready",
						Status: "True",
					},
				}, Allocatable: map[core.ResourceName]resource.Quantity{
					"cpu":    cpuQuantity,
					"memory": memoryQuantity,
					"pods":   podsQuantity,
				}, Capacity: map[core.ResourceName]resource.Quantity{
					"cpu":    cpuQuantity,
					"memory": memoryQuantity,
					"pods":   podsQuantity,
				}},
			}
		}
		for i, _ := range newNodes {
			s.nodes.Items = append(s.nodes.Items, newNodes[i])
			nodeAddEvent := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &newNodes[i]}}
			s.nodeEventChan <- nodeAddEvent
			s.nodeUpscalingChan <- newNodes[i]
		}
		return nil, nil
	}
	return changedNodes, nil
}

func (s *NodeInMemoryStorage) GetNodeUpscalingChannel() *broadcast.BroadcastServer[core.Node] {
	return s.nodeUpscalingBroadcaster
}

func (s *NodeInMemoryStorage) GetNodeDownscalingChannel() *broadcast.BroadcastServer[core.Node] {
	return s.nodeDownscalingBroadcaster
}

func (s *PodInMemoryStorage) GetPods() (core.PodList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.pods, s.podBroadcaster
}

func (s *PodInMemoryStorage) StorePods(pods core.PodList, events []metav1.WatchEvent) {
	s.pods = pods
	for _, e := range events {
		s.podEventChan <- e
	}
}

func (s *PodInMemoryStorage) DeletePods(events []metav1.WatchEvent) {
	s.pods = core.PodList{}
	for _, e := range events {
		e.Type = "DELETED"
		s.podEventChan <- e
	}
}

func (s *PodInMemoryStorage) BindPod(podIndex int, nodeName string) {
	s.pods.Items[podIndex].ObjectMeta.ResourceVersion = s.getNextResourceId()
	s.pods.Items[podIndex].Spec.NodeName = nodeName
	s.pods.Items[podIndex].Status.Phase = "Running"
	s.pods.Items[podIndex].Status.Conditions = append(s.pods.Items[podIndex].Status.Conditions, core.PodCondition{
		Type:   core.PodScheduled,
		Status: core.ConditionTrue,
	})
	s.podEventChan <- metav1.WatchEvent{
		Type:   "MODIFIED",
		Object: runtime.RawExtension{Object: &s.pods.Items[podIndex]},
	}
}

func (s *PodInMemoryStorage) FailedPod(podIndex int, status core.PodStatus) {
	s.pods.Items[podIndex].Status = status
	s.pods.Items[podIndex].Status.Phase = "Pending"
	s.pods.Items[podIndex].ObjectMeta.ResourceVersion = s.getNextResourceId()
	s.podEventChan <- metav1.WatchEvent{
		Type:   "MODIFIED",
		Object: runtime.RawExtension{Object: &s.pods.Items[podIndex]},
	}
}

func (s *PodInMemoryStorage) getNextResourceId() string {
	result := strconv.Itoa(s.nextResourceId)
	s.nextResourceId = s.nextResourceId + 1
	return result
}

func (s *NamespaceInMemoryStorage) GetNamespaces() (core.NamespaceList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.namespaces, s.namespaceBroadcaster
}

func (s *NamespaceInMemoryStorage) StoreNamespaces(namespaces core.NamespaceList) {
	s.namespaces = namespaces
}

// "Constructors"

func NewPodInMemoryStorage() PodInMemoryStorage {
	podEventChan := make(chan metav1.WatchEvent, 500)
	return PodInMemoryStorage{
		pods:           core.PodList{TypeMeta: metav1.TypeMeta{Kind: "PodList", APIVersion: "v1"}, Items: nil},
		podEventChan:   podEventChan,
		podBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "PodBroadcaster", podEventChan),
		nextResourceId: 1,
	}
}

func NewNodeInMemoryStorage() NodeInMemoryStorage {
	nodeEventChan := make(chan metav1.WatchEvent, 500)
	nodeUpscalingChan := make(chan core.Node)
	nodeDownscalingChan := make(chan core.Node)
	return NodeInMemoryStorage{
		nodes:                      core.NodeList{TypeMeta: metav1.TypeMeta{Kind: "NodeList", APIVersion: "v1"}, Items: nil},
		nodeEventChan:              nodeEventChan,
		nodeBroadcaster:            broadcast.NewBroadcastServer(context.TODO(), "NodeBroadcaster", nodeEventChan),
		nodeUpscalingChan:          nodeUpscalingChan,
		nodeDownscalingChan:        nodeDownscalingChan,
		nodeDownscalingBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "NodeDownscalingBroadcaster", nodeDownscalingChan),
		nodeUpscalingBroadcaster:   broadcast.NewBroadcastServer(context.TODO(), "NodeUpscalingBroadcaster", nodeUpscalingChan),
	}
}

func NewNamespaceInMemoryStorage() NamespaceInMemoryStorage {
	var namespace core.Namespace
	namespace.SetName("default")
	namespace.Status = core.NamespaceStatus{Phase: "Active"}
	namespaceEventChan := make(chan metav1.WatchEvent)
	return NamespaceInMemoryStorage{
		namespaces:           core.NamespaceList{TypeMeta: metav1.TypeMeta{Kind: "NamespaceList", APIVersion: "v1"}, Items: []core.Namespace{namespace}},
		namespaceEventChan:   namespaceEventChan,
		namespaceBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "NamespaceBroadcaster", namespaceEventChan),
	}
}

package debugging

import (
	"fmt"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"kube-rise/pkg/storage"
	"net/http"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
	"strconv"
)

type DebugServer struct {
	storages *storage.StorageContainer
	podCount int
	podList  core.PodList
}

func NewDebugServer(storages *storage.StorageContainer) *DebugServer {
	return &DebugServer{storages: storages, podCount: 0, podList: core.PodList{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "PodList"},
		Items:    []core.Pod{},
	}}
}

func (s *DebugServer) InitTestValues(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Example machine set
	replicas := int32(1)
	exampleMachineSet := cluster.MachineSet{
		TypeMeta: metav1.TypeMeta{APIVersion: "cluster.x-k8s.io/v1beta1", Kind: "MachineSet"},
		ObjectMeta: metav1.ObjectMeta{Name: "my-machine-set", Namespace: "kube-system", Annotations: map[string]string{
			"cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size": "0",
			"cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size": "5",
			"capacity.cluster-autoscaler.kubernetes.io/memory":            "128G",
			"capacity.cluster-autoscaler.kubernetes.io/cpu":               "16",
			"capacity.cluster-autoscaler.kubernetes.io/maxPods":           "200",
		}},
		Spec: cluster.MachineSetSpec{Replicas: &replicas, Selector: metav1.LabelSelector{MatchLabels: map[string]string{
			"machine-set-name": "my-machine-set",
		}}},
		Status: cluster.MachineSetStatus{Replicas: replicas,
			FullyLabeledReplicas: replicas,
			ReadyReplicas:        replicas,
			AvailableReplicas:    replicas,
			Conditions: []cluster.Condition{
				{
					Type:   "Ready",
					Status: "True",
				},
			},
		},
	}
	machineSetList := cluster.MachineSetList{TypeMeta: metav1.TypeMeta{Kind: "MachineSetList", APIVersion: "cluster-x.k8s.io/v1beta1"}, Items: []cluster.MachineSet{exampleMachineSet}}
	machineSetAddEvent := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &exampleMachineSet}}
	s.storages.MachineSets.StoreMachineSets(machineSetList, []metav1.WatchEvent{machineSetAddEvent})

	// Example machine
	providerid := "clusterapi://test-node"
	nodeReference := core.ObjectReference{Kind: "Node", APIVersion: "v1", Name: "test-node"}
	exampleMachine := cluster.Machine{
		TypeMeta: metav1.TypeMeta{APIVersion: "cluster.x-k8s.io/v1beta1", Kind: "Machine"},
		ObjectMeta: metav1.ObjectMeta{Name: "my-machine", Namespace: "kube-system", Labels: map[string]string{
			"machine-set-name": "my-machine-set",
		}, OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: "cluster.x-k8s.io/v1beta1",
				Kind:       "MachineSet",
				Name:       "my-machine-set",
			},
		}},
		Spec:   cluster.MachineSpec{ProviderID: &providerid},
		Status: cluster.MachineStatus{Phase: "Running", NodeRef: &nodeReference},
	}
	machineList := cluster.MachineList{TypeMeta: metav1.TypeMeta{Kind: "MachineList", APIVersion: "cluster-x.k8s.io/v1beta1"}, Items: []cluster.Machine{exampleMachine}}
	machineAddEvent := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &exampleMachine}}
	s.storages.Machines.StoreMachines(machineList, []metav1.WatchEvent{machineAddEvent})

	// Example node
	quantity, _ := resource.ParseQuantity("4")
	podQuantity, _ := resource.ParseQuantity("120")
	exampleNode := core.Node{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Node"},
		ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
		Spec:       core.NodeSpec{ProviderID: providerid},
		Status: core.NodeStatus{Phase: "Running", Conditions: []core.NodeCondition{
			{
				Type:   "Ready",
				Status: "True",
			},
		}, Allocatable: map[core.ResourceName]resource.Quantity{
			"cpu":    quantity,
			"memory": quantity,
			"pods":   podQuantity,
		}, Capacity: map[core.ResourceName]resource.Quantity{
			"cpu":    quantity,
			"memory": quantity,
			"pods":   podQuantity,
		}},
	}
	nodeList := core.NodeList{TypeMeta: metav1.TypeMeta{Kind: "NodeList", APIVersion: "v1"}, Items: []core.Node{exampleNode}}
	nodeAddEvent := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &exampleNode}}
	s.storages.Nodes.StoreNodes(nodeList, []metav1.WatchEvent{nodeAddEvent})
}

func (s *DebugServer) InitTestPods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	podsToBeCreated := r.URL.Query().Get("n")
	if podsToBeCreated == "" {
		podsToBeCreated = "1"
	}
	podsToBeCreatedNumber, _ := strconv.Atoi(podsToBeCreated)
	createdPods := make([]core.Pod, podsToBeCreatedNumber)
	for podsToBeCreatedNumber > 0 {
		s.podCount = s.podCount + 1
		cpuQuantity, _ := resource.ParseQuantity("2")
		// Example pod
		examplePod := core.Pod{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("pod-%d", s.podCount), Namespace: "default", UID: types.UID(fmt.Sprintf("pod-%d", s.podCount))},
			Spec: core.PodSpec{
				Containers: []core.Container{
					{
						Name: fmt.Sprintf("pod-%d-container", s.podCount),
						Resources: core.ResourceRequirements{
							Limits: map[core.ResourceName]resource.Quantity{
								"cpu":    cpuQuantity,
								"memory": cpuQuantity,
							},
							Requests: map[core.ResourceName]resource.Quantity{
								"cpu":    cpuQuantity,
								"memory": cpuQuantity,
							},
						},
					},
				},
				SchedulerName: "default-scheduler",
			},
			Status: core.PodStatus{
				Phase: "Pending",
				Conditions: []core.PodCondition{
					{
						Type:   core.PodScheduled,
						Status: core.ConditionFalse,
						Reason: core.PodReasonUnschedulable,
					},
				},
			},
		}
		createdPods = append(createdPods, examplePod)
		podsToBeCreatedNumber = podsToBeCreatedNumber - 1
	}
	for i, _ := range createdPods {
		s.podList.Items = append(s.podList.Items, createdPods[i])
		podCreatedEvent := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &createdPods[i]}}
		s.storages.Pods.StorePods(s.podList, []metav1.WatchEvent{podCreatedEvent})
	}
}

package inmemorystorage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	v1 "k8s.io/api/autoscaling/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kube-rise/internal/broadcast"
	"net/http"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
	"strconv"
)

// Structs

type MachineInMemoryStorage struct {
	machines           cluster.MachineList
	machineEventChan   chan metav1.WatchEvent
	machineBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
	machineCount       int
}

type MachineSetsInMemoryStorage struct {
	machineSets           cluster.MachineSetList
	machineSetsEventChan  chan metav1.WatchEvent
	nodeStorage           *NodeInMemoryStorage
	machineStorage        *MachineInMemoryStorage
	machineSetBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
}

type StatusConfigMapInMemoryStorage struct {
	statusConfigMap core.ConfigMap
}

// Implementations

func (s *MachineInMemoryStorage) GetMachines() (cluster.MachineList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.machines, s.machineBroadcaster
}

func (s *MachineInMemoryStorage) StoreMachines(ms cluster.MachineList, events []metav1.WatchEvent) {
	s.machines = ms
	for _, n := range events {
		s.machineEventChan <- n
		if n.Type == "ADDED" {
			s.machineCount = s.machineCount + 1
		}
	}
}

func (s *MachineInMemoryStorage) GetMachine(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	w.Header().Set("Content-Type", "application/json")

	// Get machine set name as path parameter
	pathParams := mux.Vars(r)
	machineName := pathParams["machineName"]
	var machineRef cluster.Machine

	for _, ms := range s.machines.Items {
		if ms.Name == machineName {
			machineRef = ms
			break
		}
	}

	err := json.NewEncoder(w).Encode(machineRef)
	if err != nil {
		fmt.Printf("Unable to encode response for get machine, error is: %v", err)
		return
	}
}

func (s *MachineInMemoryStorage) PutMachine(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	w.Header().Set("Content-Type", "application/json")

	// Get node name as path parameter
	pathParams := mux.Vars(r)
	machineName := pathParams["machineName"]

	reqBody, _ := io.ReadAll(r.Body)
	// fmt.Println(string(reqBody))
	var u cluster.Machine
	err := json.Unmarshal(reqBody, &u)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		w.WriteHeader(500)
		return
	}

	indexForReplacement := -1
	for index, machine := range s.machines.Items {
		if machine.Name == machineName {
			indexForReplacement = index
			break
		}
	}
	s.machines.Items[indexForReplacement] = u

	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		fmt.Printf("Unable to encode response for node update, error is: %v", err)
		return
	}
}

func (s *MachineInMemoryStorage) ScaleMachines(machineSet cluster.MachineSet, changedNodes []core.Node, amount int) ([]cluster.Machine, error) {
	var addedMachines []cluster.Machine
	if amount < 0 {
		// In case of downscaling we need to delete machines
		for _, changedNode := range changedNodes {
			index := -1
			for i, machine := range s.machines.Items {
				if machine.Status.NodeRef.Name == changedNode.Name {
					index = i
					break
				}
			}
			s.machineEventChan <- metav1.WatchEvent{Type: "DELETED", Object: runtime.RawExtension{Object: &s.machines.Items[index]}}
			s.machines.Items[index] = s.machines.Items[len(s.machines.Items)-1]
			s.machines.Items = s.machines.Items[:len(s.machines.Items)-1]
		}
		return nil, nil
	} else {
		providerIds := make([]string, amount)
		nodeRefs := make([]core.ObjectReference, amount)
		for amount > 0 {

			providerIds[amount-1] = "clusterapi://" + fmt.Sprintf("%s-machine-%d", machineSet.Name, s.machineCount)
			nodeRefs[amount-1] = core.ObjectReference{Kind: "Node", APIVersion: "v1", Name: fmt.Sprintf("%s-machine-%d", machineSet.Name, s.machineCount) + "-node"}
			newMachine := cluster.Machine{
				TypeMeta: metav1.TypeMeta{APIVersion: "cluster.x-k8s.io/v1beta1", Kind: "Machine"},
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-machine-%d", machineSet.Name, s.machineCount), Namespace: "kube-system", Annotations: map[string]string{
					"machine-set-name": machineSet.Name,
					"cpu":              machineSet.Annotations["capacity.cluster-autoscaler.kubernetes.io/cpu"],
					"memory":           machineSet.Annotations["capacity.cluster-autoscaler.kubernetes.io/memory"],
					"pods":             machineSet.Annotations["capacity.cluster-autoscaler.kubernetes.io/maxPods"],
				}, OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "cluster.x-k8s.io/v1beta1",
						Kind:       "MachineSet",
						Name:       machineSet.Name,
					},
				}},
				Spec:   cluster.MachineSpec{ProviderID: &providerIds[amount-1]},
				Status: cluster.MachineStatus{Phase: "Running", NodeRef: &nodeRefs[amount-1]},
			}
			s.machineCount = s.machineCount + 1
			amount = amount - 1
			addedMachines = append(addedMachines, newMachine)
		}
		for i, _ := range addedMachines {
			s.machines.Items = append(s.machines.Items, addedMachines[i])
			machineAddEvent := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &addedMachines[i]}}
			s.machineEventChan <- machineAddEvent
		}
	}
	return addedMachines, nil
}

func (s *MachineSetsInMemoryStorage) GetMachineSets() (cluster.MachineSetList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.machineSets, s.machineSetBroadcaster
}

func (s *MachineSetsInMemoryStorage) StoreMachineSets(ms cluster.MachineSetList, events []metav1.WatchEvent) {
	s.machineSets = ms
	for _, e := range events {
		s.machineSetsEventChan <- e
	}
}

func (s *MachineSetsInMemoryStorage) GetMachineSetsScale(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	w.Header().Set("Content-Type", "application/json")

	// Get machine set name as path parameter
	pathParams := mux.Vars(r)
	machineSetName := pathParams["machinesetName"]
	var machineSetRef cluster.MachineSet

	for _, ms := range s.machineSets.Items {
		if ms.Name == machineSetName {
			machineSetRef = ms
			break
		}
	}

	result := v1.Scale{TypeMeta: metav1.TypeMeta{APIVersion: "autoscaling/v1", Kind: "Scale"},
		ObjectMeta: metav1.ObjectMeta{Name: machineSetName},
		Spec:       v1.ScaleSpec{Replicas: *machineSetRef.Spec.Replicas},
		Status:     v1.ScaleStatus{Replicas: *machineSetRef.Spec.Replicas}}

	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		fmt.Printf("Unable to encode response for machineset scale, error is: %v", err)
		return
	}
}

func (s *MachineSetsInMemoryStorage) PutMachineSetsScale(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := io.ReadAll(r.Body)
	// fmt.Println(string(reqBody))
	var u v1.Scale
	err := json.Unmarshal(reqBody, &u)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		w.WriteHeader(500)
		return
	}

	// Get machine set name as path parameter
	pathParams := mux.Vars(r)
	machineSetName := pathParams["machinesetName"]
	desiredReplicas := u.Spec.Replicas
	scaleAmount := int32(0)
	machineSetIndexToChange := -1
	for i, ms := range s.machineSets.Items {
		if ms.Name == machineSetName {
			machineSetIndexToChange = i
			break
		}
	}
	scaleAmount = desiredReplicas - *s.machineSets.Items[machineSetIndexToChange].Spec.Replicas
	*s.machineSets.Items[machineSetIndexToChange].Spec.Replicas = desiredReplicas
	(*s).machineSets.Items[machineSetIndexToChange].Status.AvailableReplicas = desiredReplicas
	(*s).machineSets.Items[machineSetIndexToChange].Status.FullyLabeledReplicas = desiredReplicas
	(*s).machineSets.Items[machineSetIndexToChange].Status.ReadyReplicas = desiredReplicas
	s.machineSetsEventChan <- metav1.WatchEvent{Type: "MODIFIED", Object: runtime.RawExtension{Object: &s.machineSets.Items[machineSetIndexToChange]}}

	// We need to scale something
	if scaleAmount != 0 {
		if scaleAmount < 0 {
			// For downscaling, we first delete nodes then machines
			changedNodes, err := s.nodeStorage.ScaleNodes(nil, int(scaleAmount))
			if err != nil {
				fmt.Printf("Error when scaling down nodes. err = %s", err)
				w.WriteHeader(500)
				return
			}
			_, err = s.machineStorage.ScaleMachines(s.machineSets.Items[machineSetIndexToChange], changedNodes, int(scaleAmount))
			if err != nil {
				fmt.Printf("Error when scaling down machines. err = %s", err)
				w.WriteHeader(500)
				return
			}
		} else {
			// For upscaling, we first create machines then nodes
			addedMachines, err := s.machineStorage.ScaleMachines(s.machineSets.Items[machineSetIndexToChange], nil, int(scaleAmount))
			if err != nil {
				fmt.Printf("Error when scaling up machines. err = %s", err)
				w.WriteHeader(500)
				return
			}
			_, err = s.nodeStorage.ScaleNodes(addedMachines, int(scaleAmount))
			if err != nil {
				fmt.Printf("Error when scaling up nodes. err = %s", err)
				w.WriteHeader(500)
				return
			}
		}
	}

	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		fmt.Printf("Unable to encode response for machineset put scale, error is: %v", err)
		return
	}
}

func (s *MachineSetsInMemoryStorage) IsUpscalingPossible() bool {
	for _, ms := range s.machineSets.Items {
		maxSize, _ := strconv.Atoi(ms.Annotations["cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size"])
		if *ms.Spec.Replicas < int32(maxSize) {
			return true
		}
	}
	return false
}

func (s *MachineSetsInMemoryStorage) IsDownscalingPossible() bool {
	for _, ms := range s.machineSets.Items {
		minSize, _ := strconv.Atoi(ms.Annotations["cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"])
		if *ms.Spec.Replicas > int32(minSize) {
			return true
		}
	}
	return false
}

func (s *StatusConfigMapInMemoryStorage) GetStatusConfigMap() core.ConfigMap {
	return s.statusConfigMap
}

func (s *StatusConfigMapInMemoryStorage) StoreStatusConfigMap(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := io.ReadAll(r.Body)
	// fmt.Println(string(reqBody))
	var u core.ConfigMap
	err := json.Unmarshal(reqBody, &u)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		w.WriteHeader(500)
		return
	}

	s.statusConfigMap = u

	err = json.NewEncoder(w).Encode(s.statusConfigMap)
	if err != nil {
		fmt.Printf("Unable to encode response for node update, error is: %v", err)
		return
	}
}

// "Constructors"

func NewMachineSetInMemoryStorage(nodeStorage *NodeInMemoryStorage, machineStorage *MachineInMemoryStorage) MachineSetsInMemoryStorage {
	machineSetsEventChan := make(chan metav1.WatchEvent, 500)
	return MachineSetsInMemoryStorage{
		machineSets:           cluster.MachineSetList{TypeMeta: metav1.TypeMeta{Kind: "MachineSetList", APIVersion: "cluster-x.k8s.io/v1beta1"}, Items: nil},
		machineSetsEventChan:  machineSetsEventChan,
		nodeStorage:           nodeStorage,
		machineStorage:        machineStorage,
		machineSetBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "MachineSetBroadcaster", machineSetsEventChan),
	}
}

func NewMachineInMemoryStorage() MachineInMemoryStorage {
	machineEventChan := make(chan metav1.WatchEvent, 500)
	return MachineInMemoryStorage{
		machines:           cluster.MachineList{TypeMeta: metav1.TypeMeta{Kind: "MachineList", APIVersion: "cluster.x-k8s.io/v1beta1"}, Items: nil},
		machineEventChan:   machineEventChan,
		machineBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "MachineBroadcaster", machineEventChan),
		machineCount:       0,
	}
}

func NewStatusMapInMemoryStorage() StatusConfigMapInMemoryStorage {
	return StatusConfigMapInMemoryStorage{
		statusConfigMap: core.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "cluster-autoscaler-status", Namespace: "kube-system"}},
	}
}

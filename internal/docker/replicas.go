package docker

import "slices"

type Replica struct {
	Number int
	State  string
}

func MissingReplicas(existing []int, desired int) []int {
	missingList := []int{}
	for i := 1; i <= desired; i++ {
		if !slices.Contains(existing, i) {
			missingList = append(missingList, i)
		}
	}
	return missingList
}

func ExtraReplicas(existing []int, desired int) []int {
	extraList := []int{}
	for _, extra := range existing {
		if !(1 <= extra && extra <= desired) {
			extraList = append(extraList, extra)
		}
	}
	return extraList
}

func StoppedReplicas(replicas []Replica, desired int) []int {
	stoppedReplicasList := []int{}
	for _, replica := range replicas {
		if replica.Number <= desired && replica.State != "running" {
			stoppedReplicasList = append(stoppedReplicasList, replica.Number)
		}
	}
	slices.Sort(stoppedReplicasList)
	return stoppedReplicasList
}

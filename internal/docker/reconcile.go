package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/client"
)

func ReconcileReplicas(
	imageName, appName, networkName string,
	desiredReplicas int,
) error {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	ctx, close := context.WithTimeout(
		context.Background(),
		1*time.Minute,
	)
	defer close()

	replicasList, err := ListAppReplicasStates(appName)
	if err != nil {
		return err
	}
	replicasNumbers := make([]int, 0, len(replicasList))
	for _, replica := range replicasList {
		replicasNumbers = append(replicasNumbers, replica.Number)
	}

	extraReplicasList := ExtraReplicas(replicasNumbers, desiredReplicas)
	missingReplicasList := MissingReplicas(replicasNumbers, desiredReplicas)
	for _, replica := range extraReplicasList {
		containerName := fmt.Sprintf("%s-%d", appName, replica)

		if _, err = cli.ContainerStop(
			ctx,
			containerName,
			client.ContainerStopOptions{},
		); err != nil {
			return fmt.Errorf("stop replica %d: %w", replica, err)
		}

		if _, err = cli.ContainerRemove(
			ctx,
			containerName,
			client.ContainerRemoveOptions{},
		); err != nil {
			return fmt.Errorf("delete replica %d: %w", replica, err)
		}
	}

	if len(missingReplicasList) > 0 {
		if err := ensureImage(cli, ctx, imageName); err != nil {
			return err
		}
	}

	for _, missingReplicaID := range missingReplicasList {
		containerName := fmt.Sprintf("%s-%d", appName, missingReplicaID)

		if err := deploySingle(
			cli, ctx, imageName, containerName, networkName, 0,
		); err != nil {
			return fmt.Errorf("create replica %d: %w", missingReplicaID, err)
		}
	}

	stopped := StoppedReplicas(replicasList, desiredReplicas)
	for _, replica := range stopped {
		containerName := fmt.Sprintf("%s-%d", appName, replica)

		if err := containerStart(
			cli, ctx, containerName,
		); err != nil {
			return fmt.Errorf("start replica %d: %w", replica, err)
		}
	}

	return nil
}

package docker

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

// ListAppReplicasStates returns struct Replica
func ListAppReplicasStates(appName string) ([]Replica, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return []Replica{}, fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	ctx, close := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer close()

	result, err := cli.ContainerList(
		ctx,
		client.ContainerListOptions{
			All: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	appReplicasList := []Replica{}
	for _, c := range result.Items {
		for _, name := range c.Names {
			prefix := appName + "-"
			name = strings.TrimPrefix(name, "/")
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			replicaStr := strings.TrimPrefix(name, prefix)
			replicaID, err := strconv.Atoi(replicaStr)
			if err != nil || replicaID < 1 {
				continue
			}
			appReplicasList = append(
				appReplicasList,
				Replica{
					Number: replicaID,
					State:  string(c.State),
				})
		}
	}
	slices.SortFunc(appReplicasList, func(a, b Replica) int {
		return a.Number - b.Number
	})
	return appReplicasList, nil
}

// ListAppReplicas returns replica numbers
func ListAppReplicas(appName string) ([]int, error) {
	replicas, err := ListAppReplicasStates(appName)
	if err != nil {
		return nil, err
	}

	numbers := make([]int, 0, len(replicas))
	for _, replica := range replicas {
		numbers = append(numbers, replica.Number)
	}
	return numbers, nil
}

func deploySingle(
	cli *client.Client,
	ctx context.Context,
	imageName, containerName, networkName string,
	hostPort int,
) error {
	containerID, err := containerCreate(
		cli, ctx, imageName, containerName, networkName, hostPort,
	)
	if err != nil {
		return err
	}
	if err := containerStart(cli, ctx, containerID); err != nil {
		return err
	}

	return nil
}

func DeleteApplication(appName string, replicas int) error {
	if err := ScaleDown(appName, replicas, 0); err != nil {
		return err
	}
	return nil
}

func ShowLogs(containerName string) error {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	ctx, close := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer close()

	logs, err := cli.ContainerLogs(
		ctx,
		containerName,
		client.ContainerLogsOptions{
			ShowStderr: true,
			ShowStdout: true,
		},
	)
	if err != nil {
		return fmt.Errorf("container logs: %w", err)
	}
	defer logs.Close()

	_, err = stdcopy.StdCopy(
		os.Stdout,
		os.Stderr,
		logs,
	)
	if err != nil {
		return fmt.Errorf("stdcopy: %w", err)
	}
	return nil
}

func DeployReplicas(
	imageName, appName, networkName string,
	replicas int,
) error {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	ctx, close := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer close()

	if err = ensureImage(cli, ctx, imageName); err != nil {
		return err
	}

	for i := 1; i <= replicas; i++ {
		containerName := fmt.Sprintf("%s-%d", appName, i)
		if err := deploySingle(
			cli,
			ctx,
			imageName,
			containerName,
			networkName,
			0,
		); err != nil {
			return fmt.Errorf("create replica %d: %w", i, err)
		}
	}
	return nil
}

func ScaleUp(
	imageName, appName, networkName string,
	currentReplicas, desiredReplicas int,
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

	if err := ensureImage(cli, ctx, imageName); err != nil {
		return err
	}
	replicasList, err := ListAppReplicas(appName)
	if err != nil {
		return err
	}
	missingReplicasList := MissingReplicas(replicasList, desiredReplicas)

	for _, missingReplicaID := range missingReplicasList {
		containerName := fmt.Sprintf("%s-%d", appName, missingReplicaID)

		if err := deploySingle(
			cli, ctx, imageName, containerName, networkName, 0,
		); err != nil {
			return fmt.Errorf("create replica %d: %w", missingReplicaID, err)
		}
	}
	return nil
}

func ScaleDown(
	appName string,
	currentReplicas, desiredReplicas int,
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

	replicasList, err := ListAppReplicas(appName)
	if err != nil {
		return err
	}
	extraReplicasList := ExtraReplicas(replicasList, desiredReplicas)

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

	return nil
}

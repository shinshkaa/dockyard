package docker

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

func ListContainers() error {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	result, err := cli.ContainerList(
		ctx,
		client.ContainerListOptions{},
	)
	if err != nil {
		return fmt.Errorf("list containers: %w", err)
	}

	if len(result.Items) == 0 {
		fmt.Println("No running containers")
		return nil
	}

	for _, c := range result.Items {
		fmt.Println(c.ID[:12], c.State, c.Image)
	}
	return nil
}

func imagePull(cli *client.Client, ctx context.Context, imageName string) error {
	response, err := cli.ImagePull(
		ctx,
		imageName,
		client.ImagePullOptions{},
	)
	if err != nil {
		return fmt.Errorf("image pull: %w", err)
	}
	defer response.Close()

	if err = response.Wait(ctx); err != nil {
		return fmt.Errorf("pull image: %w", err)
	}
	return nil
}

func ensureImage(cli *client.Client, ctx context.Context, imageName string) error {
	_, err := cli.ImageInspect(
		ctx,
		imageName,
	)
	if err == nil {
		fmt.Println("Image already exists locally")
		return nil
	}
	if !errdefs.IsNotFound(err) {
		return fmt.Errorf("inspect image: %w", err)
	}
	fmt.Println("Image not found, pulling...")
	return imagePull(cli, ctx, imageName)
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

func containerCreate(
	cli *client.Client,
	ctx context.Context,
	image, name, networkName string,
	hostPort int,
) (string, error) {
	port, err := network.ParsePort("80/tcp")
	if err != nil {
		return "", fmt.Errorf("invalid container port: %w", err)
	}

	hostConfig := &container.HostConfig{}
	if hostPort > 0 {
		hostConfig.PortBindings = network.PortMap{
			port: []network.PortBinding{
				{
					HostPort: strconv.Itoa(hostPort),
				},
			},
		}
	}

	result, err := cli.ContainerCreate(
		ctx,
		client.ContainerCreateOptions{
			Image: image,
			Name:  name,
			Config: &container.Config{
				ExposedPorts: network.PortSet{
					port: {},
				},
			},
			HostConfig: hostConfig,
			NetworkingConfig: &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					networkName: {},
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("container create: %w", err)
	}
	return result.ID, nil
}

func containerStart(cli *client.Client, ctx context.Context, ID string) error {
	_, err := cli.ContainerStart(
		ctx,
		ID,
		client.ContainerStartOptions{},
	)
	if err != nil {
		return fmt.Errorf("container start: %w", err)
	}
	return nil

}

func DeployContainer(imageName, containerName, networkName string, hostPort int) error {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	if err := ensureImage(cli, ctx, imageName); err != nil {
		return err
	}

	if err := deploySingle(cli, ctx, imageName, containerName, networkName, hostPort); err != nil {
		return err
	}
	return nil
}

func StopContainer(containerName string) error {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		20*time.Second,
	)
	defer cancel()

	_, err = cli.ContainerStop(
		ctx,
		containerName,
		client.ContainerStopOptions{},
	)
	if err != nil {
		return fmt.Errorf("container stop: %w", err)
	}
	return nil
}

func DeleteContainer(containerName string) error {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return fmt.Errorf("create docker client: %w", err)
	}
	defer cli.Close()

	ctx, close := context.WithTimeout(
		context.Background(),
		20*time.Second,
	)
	defer close()

	_, err = cli.ContainerRemove(
		ctx,
		containerName,
		client.ContainerRemoveOptions{},
	)
	if err != nil {
		return fmt.Errorf("remove container: %w", err)
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

func ListAppReplicas(appName string) ([]int, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return []int{}, fmt.Errorf("create docker client: %w", err)
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

	activeContainersList := []int{}
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
			activeContainersList = append(activeContainersList, replicaID)
		}
	}
	slices.Sort(activeContainersList)
	return activeContainersList, nil
}

func CreateNetwork(networkName string) error {
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

	_, err = cli.NetworkCreate(
		ctx,
		networkName,
		client.NetworkCreateOptions{
			Driver: "bridge",
		},
	)
	if err != nil {
		return fmt.Errorf("network creation: %w", err)
	}
	return nil
}

func ConnectNetwork(networkName, containerName string) error {
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

	if _, err := cli.NetworkConnect(
		ctx,
		networkName,
		client.NetworkConnectOptions{
			Container: containerName,
		},
	); err != nil {
		return fmt.Errorf("network connection: %w", err)
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

	for i := currentReplicas; i >= desiredReplicas+1; i-- {
		containerName := fmt.Sprintf("%s-%d", appName, i)

		if _, err = cli.ContainerStop(
			ctx,
			containerName,
			client.ContainerStopOptions{},
		); err != nil {
			return fmt.Errorf("stop replica %d: %w", i, err)
		}

		if _, err = cli.ContainerRemove(
			ctx,
			containerName,
			client.ContainerRemoveOptions{},
		); err != nil {
			return fmt.Errorf("delete replica %d: %w", i, err)
		}
	}
	return nil
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

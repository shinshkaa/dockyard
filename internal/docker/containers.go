package docker

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

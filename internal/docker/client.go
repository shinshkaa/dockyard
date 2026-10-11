package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/api/types/container"
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

func containerCreate(cli *client.Client, ctx context.Context, image, name string) (string, error) {
	result, err := cli.ContainerCreate(
		ctx,
		client.ContainerCreateOptions{
			Image: image,
			Name:  name,
			Config: &container.Config{
				Cmd: []string{"sleep", "3600"},
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

func DeployContainer(imageName, containerName string) error {
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

	containerID, err := containerCreate(cli, ctx, imageName, containerName)
	if err != nil {
		return err
	}
	err = containerStart(cli, ctx, containerID)
	if err != nil {
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
package docker

import (
	"context"
	"fmt"
	"time"

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

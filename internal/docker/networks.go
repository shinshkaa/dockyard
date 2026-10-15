package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/client"
)

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

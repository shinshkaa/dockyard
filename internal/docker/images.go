package docker

import (
	"context"
	"fmt"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

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

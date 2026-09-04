package main

import (
	"context"
	"fmt"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

type dockerRuntime struct {
	client *client.Client
}

func (d *dockerRuntime) Create(ctx context.Context) (containerInfo, error) {
	response, err := d.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: "nginx:alpine",
	})
	if err != nil {
		return containerInfo{}, err
	}
	return containerInfo{ID: response.ID}, nil
}

func (d *dockerRuntime) Start(ctx context.Context, id string) error {
	_, err := d.client.ContainerStart(ctx, id, client.ContainerStartOptions{})
	return err
}

func (d *dockerRuntime) Inspect(ctx context.Context, id string) (containerInfo, error) {
	response, err := d.client.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return containerInfo{}, fmt.Errorf("%w: %s", errContainerNotFound, id)
		}
		return containerInfo{}, err
	}
	return containerInfo{
		ID:     response.Container.ID,
		Status: string(response.Container.State.Status),
	}, nil
}

func (d *dockerRuntime) Stop(ctx context.Context, id string) error {
	_, err := d.client.ContainerStop(ctx, id, client.ContainerStopOptions{})
	return err
}

func (d *dockerRuntime) Remove(ctx context.Context, id string) error {
	_, err := d.client.ContainerRemove(ctx, id, client.ContainerRemoveOptions{})
	return err
}

package main

import (
	"context"
	"fmt"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

const (
	managedLabel   = "secure-browser-poc.managed"
	sessionIDLabel = "secure-browser-poc.session-id"
)

type dockerRuntime struct {
	client *client.Client
}

func (d *dockerRuntime) Create(ctx context.Context, sessionID string) (containerInfo, error) {
	response, err := d.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image: "nginx:alpine",
			Labels: map[string]string{
				managedLabel:   "true",
				sessionIDLabel: sessionID,
			},
		},
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

func (d *dockerRuntime) FindBySessionID(ctx context.Context, sessionID string) (containerInfo, error) {
	filters := client.Filters{}
	filters = filters.Add("label", sessionIDLabel+"="+sessionID)
	response, err := d.client.ContainerList(ctx, client.ContainerListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return containerInfo{}, err
	}

	matches := make([]string, 0, len(response.Items))
	for _, item := range response.Items {
		if item.Labels[managedLabel] == "true" && item.Labels[sessionIDLabel] == sessionID {
			matches = append(matches, item.ID)
		}
	}
	if len(matches) == 0 {
		return containerInfo{}, fmt.Errorf("%w: %s", errContainerNotFound, sessionID)
	}
	if len(matches) > 1 {
		return containerInfo{}, fmt.Errorf("multiple managed containers found for session %s", sessionID)
	}

	return d.inspectByContainerID(ctx, matches[0])
}

func (d *dockerRuntime) inspectByContainerID(ctx context.Context, id string) (containerInfo, error) {
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

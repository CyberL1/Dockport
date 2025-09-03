package utils

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func BootOfflineContainer(containerName string) error {
	cli, _ := client.NewClientWithOpts(client.FromEnv)

	_, err := cli.Ping(context.Background())
	if err != nil {
		fmt.Println("Docker socket not connected, cannot check if container is offline")
		return nil
	}

	containers, _ := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", "^"+containerName+"$")),
		All:     true,
	})

	if len(containers) == 1 && containers[0].State == container.StateExited {
		fmt.Println("container offline, booting...")
		cli.ContainerStart(context.Background(), containers[0].ID, container.StartOptions{})
	}
	return nil
}

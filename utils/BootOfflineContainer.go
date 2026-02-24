package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func BootOfflineContainer(containerName string) {
	cli, _ := client.NewClientWithOpts(client.FromEnv)

	_, err := cli.Ping(context.Background())
	if err != nil {
		fmt.Println("Docker socket not connected, cannot check if container is offline")
		return
	}

	containers, _ := cli.ContainerList(context.Background(), container.ListOptions{All: true})

	for _, c := range containers {
		if strings.EqualFold(containerName, strings.TrimPrefix(c.Names[0], "/")) && c.State == container.StateExited {
			fmt.Println("container offline, booting...")
			cli.ContainerStart(context.Background(), c.ID, container.StartOptions{})
			break
		}
	}
}

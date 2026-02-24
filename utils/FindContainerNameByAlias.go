package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func FindContainerNameByAlias(containerName string) string {
	cli, _ := client.NewClientWithOpts(client.FromEnv)

	_, err := cli.Ping(context.Background())
	if err != nil {
		fmt.Println("Docker socket not connected, cannot find container by alias")
		return containerName
	}

	containersWithAlias, _ := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", "Dockport.alias")),
		All:     true,
	})

	for _, c := range containersWithAlias {
		if strings.EqualFold(containerName, c.Labels["Dockport.alias"]) {
			containerName = strings.TrimPrefix(c.Names[0], "/")
			break
		}
	}

	return containerName
}

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

	containersByAlias, _ := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", "Dockport.alias="+strings.ToLower(containerName))),
		All:     true,
	})

	if len(containersByAlias) < 1 {
		return containerName
	}

	var aliasedContainerName string
	for _, container := range containersByAlias {
		aliasedContainerName = strings.TrimPrefix(container.Names[0], "/")
	}
	return aliasedContainerName
}

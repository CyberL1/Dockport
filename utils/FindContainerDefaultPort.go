package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func FindContainerDefaultPort(containerName string) int {
	cli, _ := client.NewClientWithOpts(client.FromEnv)

	_, err := cli.Ping(context.Background())
	if err != nil {
		fmt.Println("Docker socket not connected, cannot find container's defualt port")
		return 80
	}

	containersWithPortLabel, _ := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", "Dockport.port")),
		All:     true,
	})

	containerDefaultPort := 80
	for _, c := range containersWithPortLabel {
		if strings.EqualFold(containerName, strings.TrimPrefix(c.Names[0], "/")) {
			containerDefaultPort, err = strconv.Atoi(c.Labels["Dockport.port"])
			if err != nil {
				fmt.Println("Failed to parse port, defaulting to 80")
				containerDefaultPort = 80
				continue
			}
			break
		}
	}
	return containerDefaultPort
}

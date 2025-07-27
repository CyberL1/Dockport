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
		Filters: filters.NewArgs(filters.Arg("label", "Dockport.port"))})

	var containerDefaultPort int
	for _, container := range containersWithPortLabel {
		cleanName := strings.TrimPrefix(container.Names[0], "/")

		if strings.EqualFold(cleanName, containerName) || containerName == container.Labels["Dockport.alias"] {
			containerDefaultPort, err = strconv.Atoi(container.Labels["Dockport.port"])
			if err != nil {
				fmt.Println("Failed to parse port, defaulting to 80")
				containerDefaultPort = 80
				continue
			}
			break
		}

		// Set port if container is missing the label
		containerDefaultPort = 80
	}

	return containerDefaultPort
}

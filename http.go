package main

import (
	"dockport/utils"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func startHTTPProxy(proxyDomain string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostname := strings.TrimSuffix(r.Host, "."+proxyDomain)
		subdomains := strings.Split(hostname, ".")

		var containerName string
		var containerPort int
		var err error

		rootContainer := strings.TrimSpace(os.Getenv("HTTP_ROOT_CONTAINER"))

		if hostname == proxyDomain {
			if rootContainer == "" {
				fmt.Fprint(w, "Set HTTP_ROOT_CONTAINER environment variable to use this page.")
				return
			}

			containerName = utils.FindContainerNameByAlias(rootContainer)
			containerPort = utils.FindContainerDefaultPort(containerName)
		} else {
			containerName = subdomains[len(subdomains)-1]

			if strings.Contains(containerName, "-") {
				containerNameSplitted := strings.Split(containerName, "-")
				containerPort, err = strconv.Atoi(containerNameSplitted[len(containerNameSplitted)-1])
				if err == nil {
					containerName = utils.FindContainerNameByAlias(strings.Join(containerNameSplitted[:len(containerNameSplitted)-1], "-"))
				} else {
					containerPort = utils.FindContainerDefaultPort(containerName)
				}
			} else {
				containerName = utils.FindContainerNameByAlias(containerName)
				containerPort = utils.FindContainerDefaultPort(containerName)
			}
		}

		containerAddress := fmt.Sprintf("http://%s:%d", containerName, containerPort)
		containerAddressParsed, _ := url.Parse(containerAddress)

		proxy := httputil.NewSingleHostReverseProxy(containerAddressParsed)

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Println("http proxy error:", err)

			if os.Getenv("BOOT_OFFLINE_CONTAINERS") == "true" && strings.HasSuffix(err.Error(), "no such host") {
				utils.BootOfflineContainer(containerName)

				for {
					_, err := http.Get(containerAddress)
					if err == nil {
						proxy.ServeHTTP(w, r)
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
		proxy.ServeHTTP(w, r)
	})

	http.ListenAndServe(":80", handler)
}

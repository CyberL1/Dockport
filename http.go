package main

import (
	"dockport/utils"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

func startHTTPProxy(proxyDomain string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostname := strings.TrimSuffix(r.Host, "."+proxyDomain)
		subdomains := strings.Split(hostname, ".")

		containerName := subdomains[len(subdomains)-1]
		containerNameSplitted := strings.Split(containerName, "-")

		containerPort, err := strconv.Atoi(containerNameSplitted[len(containerNameSplitted)-1])

		if err == nil {
			containerName = strings.Join(containerNameSplitted[:len(containerNameSplitted)-1], "-")
		} else {
			containerPort = utils.FindContainerDefaultPort(containerName)
		}

		containerName = utils.FindContainerNameByAlias(containerName)
		containerAddress, _ := url.Parse(fmt.Sprintf("http://%s:%d", containerName, containerPort))

		httputil.NewSingleHostReverseProxy(containerAddress).ServeHTTP(w, r)
	})

	http.ListenAndServe(":80", handler)
}

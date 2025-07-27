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
		parts := strings.TrimSuffix(r.Host, "."+proxyDomain)
		partsSplitted := strings.Split(parts, "-")

		containerName := parts
		containerPort, err := strconv.Atoi(partsSplitted[len(partsSplitted)-1])

		if err == nil {
			containerName = strings.Join(partsSplitted[:len(partsSplitted)-1], "-")
		} else {
			containerPort = utils.FindContainerDefaultPort(containerName)
		}

		containerName = utils.FindContainerNameByAlias(containerName)
		containerAddress, _ := url.Parse(fmt.Sprintf("http://%s:%d", containerName, containerPort))

		httputil.NewSingleHostReverseProxy(containerAddress).ServeHTTP(w, r)
	})

	http.ListenAndServe(":80", handler)
}

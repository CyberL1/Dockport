package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyDomain := os.Getenv("PROXY_DOMAIN")
		if proxyDomain == "" {
			fmt.Println("Error: PROXY_DOMAIN environment variable not set")
			os.Exit(1)
		}

		parts := strings.TrimSuffix(r.Host, "."+proxyDomain)
		partsSplitted := strings.Split(parts, "-")

		containerName := parts
		containerPort, err := strconv.Atoi(partsSplitted[len(partsSplitted)-1])

		if err == nil {
			containerName = strings.Join(partsSplitted[:len(partsSplitted)-1], "-")
		} else {
			containerPort = 80
		}

		containerAddress, _ := url.Parse(fmt.Sprintf("http://%s:%d", containerName, containerPort))
		httputil.NewSingleHostReverseProxy(containerAddress).ServeHTTP(w, r)
	})

	fmt.Println("Dockport started")
	http.ListenAndServe(":80", handler)
}

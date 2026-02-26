package main

import (
	"fmt"
	"os"
)

func main() {
	proxyDomain := os.Getenv("PROXY_DOMAIN")
	if proxyDomain == "" {
		fmt.Println("Error: PROXY_DOMAIN environment variable not set")
		os.Exit(1)
	}

	_, err := os.Stat("/var/run/docker.sock")
	if err != nil {
		fmt.Println("Docker socket not mounted, some features relying on the socket won't be avaliable")
	}

	// Check for necessery directories
	if _, err := os.Stat("data"); err != nil {
		os.Mkdir("data", 0755)
	}

	go startHTTPProxy(proxyDomain)
	go startSSHProxy(proxyDomain)

	// Block
	select {}
}

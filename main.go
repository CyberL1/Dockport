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

	go startHTTPProxy(proxyDomain)
	go startSSHProxy()

	// Block
	select {}
}

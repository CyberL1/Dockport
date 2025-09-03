package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"dockport/utils"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func startSSHProxy() {
	keyBytes, err := os.ReadFile("data/ssh_key.pem")
	if err != nil {
		fmt.Println("Server key not found, generating...")

		key, _ := rsa.GenerateKey(rand.Reader, 2048)

		der := x509.MarshalPKCS1PrivateKey(key)
		block := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: der,
		}

		os.WriteFile("data/ssh_key.pem", pem.EncodeToMemory(block), 0600)
		keyBytes, _ = os.ReadFile("data/ssh_key.pem")
	}

	key, err := ssh.ParsePrivateKey([]byte(keyBytes))
	if err != nil {
		fmt.Println("Failed to parse private key:", err)
	}

	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	config.AddHostKey(key)

	listener, err := net.Listen("tcp", ":22")
	if err != nil {
		fmt.Println("Failed to listen on :22:", err)
	}
	fmt.Println("SSH jump host listening on port 22...")

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept incoming connection:", err)
			continue
		}

		go func(nConn net.Conn, config *ssh.ServerConfig) {
			defer nConn.Close()

			sshConn, chans, reqs, err := ssh.NewServerConn(nConn, config)
			if err != nil {
				fmt.Println("SSH handshake failed:", err)
				return
			}
			defer sshConn.Close()

			go ssh.DiscardRequests(reqs)

			for newChannel := range chans {
				if newChannel.ChannelType() != "direct-tcpip" {
					newChannel.Reject(ssh.ConnectionFailed, "do not connect to this host directly, use it as a jump host.")
					continue
				}

				var channelData struct {
					DestAddr string
					DestPort uint32
					OrigAddr string
					OrigPort uint32
				}

				if err := ssh.Unmarshal(newChannel.ExtraData(), &channelData); err != nil {
					fmt.Println("Failed to parse direct-tcpip data:", err)
					newChannel.Reject(ssh.ConnectionFailed, "bad direct-tcpip request")
					return
				}

				// Check if DestAddr is a container alias and replace it with the actual container name, otherwise use it as is
				channelData.DestAddr = utils.FindContainerNameByAlias(channelData.DestAddr)

				dest := fmt.Sprintf("%s:%d", channelData.DestAddr, channelData.DestPort)
				fmt.Printf("Proxying direct-tcpip request to %s\n", dest)

				targetConn, err := net.Dial("tcp", dest)
				if err != nil {
					fmt.Printf("Failed to connect to destination %s:\n %v", dest, err)

					// Boot container if it is offline
					if os.Getenv("BOOT_OFFLINE_CONTAINERS") == "true" {
						utils.BootOfflineContainer(channelData.DestAddr)

						for {
							targetConn, err = net.Dial("tcp", dest)
							if err == nil {
								break
							}
							time.Sleep(100 * time.Millisecond)
						}
					} else {
						newChannel.Reject(ssh.ConnectionFailed, err.Error())
						return
					}
				}

				channel, requests, err := newChannel.Accept()
				if err != nil {
					fmt.Println("Could not accept channel:", err)
					targetConn.Close()
					return
				}

				go ssh.DiscardRequests(requests)

				go func() {
					defer targetConn.Close()
					defer channel.Close()

					io.Copy(targetConn, channel)
				}()
				go func() {
					defer targetConn.Close()
					defer channel.Close()

					io.Copy(channel, targetConn)
				}()
			}
		}(clientConn, config)
	}
}

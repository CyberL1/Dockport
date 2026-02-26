package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"dockport/utils"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
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

			if containerName == "a" {
				if r.URL.Path == "/Dockport.cer" {
					file, err := os.ReadFile("data/tls/tls.cer")
					if errors.Is(err, os.ErrNotExist) {
						fmt.Fprintln(w, "file not found")
						return
					}

					fmt.Fprint(w, string(file))
				}
				return
			}

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
		containerAddressParsed, err := url.Parse(containerAddress)
		if err != nil {
			fmt.Println(err)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(containerAddressParsed)

		TIMEOUT_INTERVAL_SECONDS := 1
		TIMEOUT_MAX_RETRIES := 5

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Println("http proxy error:", err)

			if strings.HasSuffix(err.Error(), "no such host") {
				if os.Getenv("BOOT_OFFLINE_CONTAINERS") == "true" {
					utils.BootOfflineContainer(containerName)

					for range TIMEOUT_MAX_RETRIES {
						_, err := http.Get(containerAddress)
						if err == nil {
							proxy.ServeHTTP(w, r)
							return
						}
						time.Sleep(time.Duration(TIMEOUT_INTERVAL_SECONDS) * time.Second)
					}
				}
				fmt.Fprintln(w, "Could not reach the container")
			}
		}
		proxy.ServeHTTP(w, r)
	})

	if os.Getenv("ENABLE_HTTPS") == "true" {
		var caCert *x509.Certificate
		var caKey *ecdsa.PrivateKey

		if _, err := os.Stat("data/tls/tls.cer"); errors.Is(err, os.ErrNotExist) {
			caKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
			if err != nil {
				fmt.Println(err)
				return
			}

			template := &x509.Certificate{
				SerialNumber:          big.NewInt(time.Now().UnixNano()),
				Subject:               pkix.Name{Organization: []string{"Dockport CA"}},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(1, 0, 0),
				KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
				IsCA:                  true,
				BasicConstraintsValid: true,
			}

			certDer, err := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
			if err != nil {
				fmt.Println(err)
				return
			}
			certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})

			caCert, err = x509.ParseCertificate(certDer)
			if err != nil {
				fmt.Println(err)
				return
			}

			os.WriteFile("data/tls/tls.cer", certPem, 0644)

			priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
			if err != nil {
				fmt.Println(err)
				return
			}

			template = &x509.Certificate{
				SerialNumber: big.NewInt(time.Now().UnixNano()),
				NotBefore:    time.Now(),
				NotAfter:     time.Now().AddDate(1, 0, 0),
			}

			certDer, _ = x509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, caKey)

			certPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})
			keyBytes, err := x509.MarshalECPrivateKey(caKey)
			if err != nil {
				fmt.Println(err)
				return
			}

			keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

			os.WriteFile("data/tls/tls.crt", certPem, 0644)
			os.WriteFile("data/tls/tls.key", keyPem, 0600)
		}

		// Listen on port 80 for redirect
		go func() {
			http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
			}))
		}()

		server := http.Server{
			Addr:    ":443",
			Handler: handler,
			TLSConfig: &tls.Config{
				GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
					domain := chi.ServerName
					if domain == "" {
						return nil, errors.New("no server name provided")
					}

					certPath := filepath.Join("data/domains/" + domain + ".crt")
					keyPath := filepath.Join("data/domains/" + domain + ".key")

					if _, err := os.Stat(certPath); err == nil {
						cert, err := tls.LoadX509KeyPair(certPath, keyPath)
						if err != nil {
							return nil, err
						}

						return &cert, nil
					}

					if caCert == nil {
						var err error

						certPem, _ := os.ReadFile("data/tls/tls.cer")
						block, _ := pem.Decode(certPem)
						if block == nil || block.Type != "CERTIFICATE" {
							return nil, errors.New("invalid CA certificate PEM")
						}

						caCert, err = x509.ParseCertificate(block.Bytes)
						if err != nil {
							return nil, err
						}
					}

					if caKey == nil {
						caKeyBytes, err := os.ReadFile("data/tls/tls.key")
						if err != nil {
							return nil, err
						}

						block, _ := pem.Decode(caKeyBytes)
						if block == nil || block.Type != "EC PRIVATE KEY" {
							return nil, errors.New("invalid CA key PEM")
						}

						caKey, err = x509.ParseECPrivateKey(block.Bytes)
						if err != nil {
							return nil, err
						}
					}

					template := &x509.Certificate{
						SerialNumber: big.NewInt(time.Now().UnixNano()),
						NotBefore:    time.Now(),
						NotAfter:     time.Now().AddDate(1, 0, 0),
						DNSNames:     []string{domain},
					}

					priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
					if err != nil {
						return nil, err
					}

					certDer, err := x509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, caKey)
					if err != nil {
						return nil, err
					}

					keyBytes, err := x509.MarshalECPrivateKey(priv)
					if err != nil {
						return nil, err
					}

					certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})
					keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

					cert, err := tls.X509KeyPair(certPem, keyPem)
					if err != nil {
						return nil, err
					}

					os.WriteFile(certPath, certPem, 0644)
					os.WriteFile(keyPath, keyPem, 0600)

					return &cert, nil
				},
			},
		}

		if err := server.ListenAndServeTLS("", ""); err != nil {
			fmt.Println(err)
		}
	} else {
		http.ListenAndServe(":80", handler)
	}
}

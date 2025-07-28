# What is this?

Dockport is a lighweight, automatic proxy for docker containers.

# How to set this up?

1. Create a network for Dockport:
  ```bash
  docker network create Dockport
  ```

2. Run the proxy using:
  ```bash
  docker run -d --name Dockport -v Dockport:/Dockport/data -e PROXY_DOMAIN=localhost --network Dockport -p 80:80 -p 2222:22 ghcr.io/cyberl1/dockport
  ```

# How to use this for http?

1. Run another container connected to the proxy network:
  ```bash
  docker run -d --name another-container --network Dockport nginx:alpine
  ```

2. Go to http://another-container.localhost in your browser and see the result. If your container is listening on diffrent port (i.e 8080), then go to http://another-container-8080.localhost

# How to use this for ssh?

1. Run another container connected to the proxy network:
  ```bash
  docker run -d --name another-container --network Dockport alpine /bin/sh -c 'apk update && apk add openssh && adduser user --gecos "" --disabled-password && echo "user:password" | chpasswd && ssh-keygen -A && /usr/sbin/sshd -D'
  ```

2. Connect to it using:
  ```bash
  ssh -J localhost:2222 user@another-container
  ```

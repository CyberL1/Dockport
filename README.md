# What is this?

Dockport is a lighweight, automatic proxy for docker containers.

# How to use this?

1. Create a network for Dockport:
  ```bash
  docker network create Dockport
  ```

2. Run the proxy using:
   ```bash
   docker run -d --name Dockport -v Dockport:/Dockport/data -e PROXY_DOMAIN=localhost --network Dockport -p 80:80 -v 2222:22 ghcr.io/cyberl1/dockport
   ```

3. Run another container connected to the proxy network:
  ```bash
  docker run -d --name another-container --network Dockport nginx:alpine
  ```

4. Go to http://another-container.localhost in your browser and see the result. If your container is listening on diffrent port (i.e 8080), then go to http://another-container-8080.localhost

# Docker-compose basic example

In this section we quickly go over a basic docker-compose file exposing a simple service using the docker provider.  
This will also be used as a starting point for the other docker-compose guides.  

## Setup

- Edit a `docker-compose.yml` file with the following content:

```yaml
--8<-- "content/user-guides/docker-compose/basic-example/docker-compose.yml"
```

- Replace `whoami.localhost` by your **own domain** within the `traefik.http.routers.whoami.rule` label of the `whoami` service.
- Run `docker-compose up -d` within the folder where you created the previous file.
- Wait a bit and visit `http://your_own_domain` to confirm everything went fine.
	You should see the output of the whoami service. Something similar to:
	
	```text
	Hostname: d7f919e54651
	IP: 127.0.0.1
	IP: 192.168.64.2
	GET / HTTP/1.1
	Host: whoami.localhost
	User-Agent: curl/7.52.1
	Accept: */*
	Accept-Encoding: gzip
	X-Forwarded-For: 192.168.64.1
	X-Forwarded-Host: whoami.localhost
	X-Forwarded-Port: 80
	X-Forwarded-Proto: http
	X-Forwarded-Server: 7f0c797dbc51
	X-Real-Ip: 192.168.64.1
	```

## Details

- As an example we use [whoami](https://github.com/traefik/whoami) (a tiny Go server that prints os information and HTTP request to output) which was used to define our `simple-service` container.

- We define an entry point, along with the exposure of the matching port within docker-compose, which basically allow us to "open and accept" HTTP traffic: 

```yaml
command:
  # Traefik will listen to incoming request on the port 80 (HTTP)
  - "--entrypoints.web.address=:80"

ports:
  - "80:80"
```

- We expose the Traefik API to be able to check the configuration if needed:

```yaml
command:
  # Traefik will listen on port 8080 by default for API request.
  - "--api.insecure=true"

ports:
  - "8080:8080"
```

!!! Note

    If you are working on a remote server, you can use the following command to display configuration (require `curl` & `jq`): 
     
    ```bash
    curl -s 127.0.0.1:8080/api/rawdata | jq .
    ```

- We allow Traefik to gather configuration from Docker:

```yaml
traefik:
  command:
    # Enabling docker provider
    - "--providers.docker=true"
    # Do not expose containers unless explicitly told so
    - "--providers.docker.exposedbydefault=false"
  volumes:
    - "/var/run/docker.sock:/var/run/docker.sock:ro"

whoami:
  labels:
    # Explicitly tell Traefik to expose this container
    - "traefik.enable=true"
    # The domain the service will respond to
    - "traefik.http.routers.whoami.rule=Host(`whoami.localhost`)"
    # Allow request only from the predefined entry point named "web"
    - "traefik.http.routers.whoami.entrypoints=web"
```

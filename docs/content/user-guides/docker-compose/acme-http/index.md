---
title: "Traefik Docker HTTP Challenge Documentation"
description: "Learn how to create a certificate with the Let's Encrypt HTTP challenge to use HTTPS on a Service exposed with Traefik Proxy. Read the technical documentation."
---

# Docker-compose with Let's Encrypt : HTTP Challenge

This guide aims to demonstrate how to create a certificate with the Let's Encrypt HTTP challenge to use https on a simple service exposed with Traefik.  
Please also read the [basic example](../basic-example) for details on how to expose such a service.  

## Prerequisite

For the HTTP challenge you will need:

- A publicly accessible host allowing connections on port `80` & `443` with docker & docker-compose installed.
- A DNS record with the domain you want to expose pointing to this host.

## Setup

- Create a `docker-compose.yml` on your remote server with the following content:

```yaml
--8<-- "content/user-guides/docker-compose/acme-http/docker-compose.yml"
```

- Replace `postmaster@example.com` by your **own email** within the `certificatesresolvers.myresolver.acme.email` command line argument of the `traefik` service.
- Replace `whoami.example.com` by your **own domain** within the `traefik.http.routers.whoami.rule` label of the `whoami` service.
- Optionally uncomment the following lines if you want to test/debug:

	```yaml
	#- "--log.level=DEBUG"
	#- "--certificatesresolvers.myresolver.acme.caserver=https://acme-staging-v02.api.letsencrypt.org/directory"
	```

- Run `docker compose up -d` within the folder where you created the previous file.
- Wait a bit and visit `https://your_own_domain` to confirm everything went fine.

!!! Note

    If you uncommented the `acme.caserver` line, you will get an SSL error, but if you display the certificate and see it was emitted by `Fake LE Intermediate X1` then it means all is good.
    (It is the staging environment intermediate certificate used by Let's Encrypt).  
   You can now safely comment the `acme.caserver` line, remove the `letsencrypt/acme.json` file and restart Traefik to issue a valid certificate.

## Explanation

What changed between the basic example:

- We configure a second entry point for the HTTPS traffic:

```yaml
command:
  # Traefik will listen to incoming request on the port 443 (https)
  - "--entryPoints.websecure.address=:443"
ports:
  - "443:443"
```

- We configure the HTTP Let's Encrypt challenge:

```yaml
command:
  # Enable a http challenge named "myresolver"
  - "--certificatesresolvers.myresolver.acme.httpchallenge=true"
  # Tell it to use our predefined entrypoint named "web"
  - "--certificatesresolvers.myresolver.acme.httpchallenge.entrypoint=web"
  # The email to provide to Let's Encrypt
  - "--certificatesresolvers.myresolver.acme.email=postmaster@example.com"
```

- We add a volume to store our certificates:

```yaml
volumes:
  # Create a letsencrypt dir within the folder where the docker-compose file is
  - "./letsencrypt:/letsencrypt"

command:
  # Tell to store the certificate on a path under our volume
  - "--certificatesresolvers.myresolver.acme.storage=/letsencrypt/acme.json"
```

- We configure the `whoami` service to tell Traefik to use the certificate resolver named `myresolver` we just configured:

```yaml
labels:
  # Uses the Host rule to define which certificate to issue
  - "traefik.http.routers.whoami.tls.certresolver=myresolver"
```

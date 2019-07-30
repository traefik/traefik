# Docker-compose with let's encrypt : HTTP Challenge

This guide aim to demonstrate how to create a certificate with the let's encrypt HTTP challenge to use https on a simple service exposed with Traefik.  
Please also read the [Docker-compose : Basic Example](../basic-example) for details on how to expose such a service.  

## Prerequisite

For the Http challenge you will need :

- A publicly accessible host allowing connections on port 80 & 443 with docker & docker-compose installed.
- A DNS record with the domain you want to expose pointing to this host.

## Setup

- Create a *docker-compose.yml* on your remote server with the following content :

```yaml
--8<-- "content/user-guides/docker-compose/acme-http/docker-compose.yml"
```

- Replace `postmaster@mydomain.com` by your **own email** within the *certificatesresolvers.myhttpchallenge.acme.email* command line argument of the *traefik* service.
- Replace `whoami.mydomain.com` by your **own domain** within the *traefik.http.routers.whoami.rule* label of the *whoami* service.
- Optionnaly uncomment the following lines if you want to test / debug :

	```text
	#- "--log.level=DEBUG"
	#- "--certificatesresolvers.myhttpchallenge.acme.caserver=https://acme-staging-v02.api.letsencrypt.org/directory"
	```

- Run `docker-compose up -d` within the folder where you created the previous file.
- Wait a bit and visit `https://your_own_domain` to confirm everything went fine.

> If you uncommented the *acme.caserver* line, you will get an SSL error, but if you display the certificate and see it was emitted by *Fake LE Intermediate X1* then it means all is good. (It is the staging environment intermediate certificate used by let's encrypt).  
You can now safely comment the *acme.caserver* line, remove the *letsencrypt/acme.json* file and restart Traefik to issue a valid certificate.

## Explanation

What changed between the basic example :

- We configure a second entrypoint for the https traffic :

```yaml
command:
  - "--entrypoints.websecure.address=:443" # Traefik will listen to incoming request on the port 443 (https)
ports:
  - "0.0.0.0:443:443" # We allow any http requests from any ips to reach our Traefik container on the 443 port
```

- We configure the Https let's encrypt challenge :

```yaml
command:
	- "--certificatesresolvers.myhttpchallenge.acme.httpchallenge=true" # Enable a http challenge named "myhttpchallenge"
	- "--certificatesresolvers.myhttpchallenge.acme.httpchallenge.entrypoint=web" # Tell it to use our predefined entrypoint named "web"
	- "--certificatesresolvers.myhttpchallenge.acme.email=postmaster@mydomain.com" # The email to provide to let's encrypt
```

- We add a volume to store our certificates :

```yaml
volumes:
	- "./letsencrypt:/letsencrypt" # Create a letsencrypt dir within the folder where the docker-compose file is

command:
	- "--certificatesresolvers.myhttpchallenge.acme.storage=/letsencrypt/acme.json" # Tell to store the certificate on a path under our volume
```

- We configure the *whoami* service to tell Traefik to use the certresolver named "myhttpchallenge" we just configured :

```yaml
labels:
	- "traefik.http.routers.whoami.tls.certresolver=myhttpchallenge" # Uses the Host rule to define which certificate to issue
```
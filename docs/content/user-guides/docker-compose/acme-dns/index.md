---
title: "Traefik Docker DNS Challenge Documentation"
description: "Learn how to create a certificate with the Let's Encrypt DNS challenge to use HTTPS on a Service exposed with Traefik Proxy. Read the tehnical documentation."
---

# Docker-compose with let's encrypt: DNS Challenge

This guide aim to demonstrate how to create a certificate with the let's encrypt DNS challenge to use https on a simple service exposed with Traefik.  
Please also read the [basic example](../basic-example) for details on how to expose such a service.  

## Prerequisite

For the DNS challenge, you'll need:

- A working [provider](../../../https/acme.md#providers) along with the credentials allowing to create and remove DNS records.  

!!! info "Variables may vary depending on the Provider."
	 Please note this guide may vary depending on the provider you use.
	 The only things changing are the names of the variables you will need to define in order to configure your provider so it can create DNS records.  
	 Please refer the [list of providers](../../../https/acme.md#providers) given right above and replace all the environment variables with the ones described in this documentation.

## Setup

- Create a `docker-compose.yml` file with the following content:

```yaml
--8<-- "content/user-guides/docker-compose/acme-dns/docker-compose.yml"
```

- Replace the environment variables by your own:
    
    ```yaml
    environment:
      - "OVH_ENDPOINT=[YOUR_OWN_VALUE]"
      - "OVH_APPLICATION_KEY=[YOUR_OWN_VALUE]"
      - "OVH_APPLICATION_SECRET=[YOUR_OWN_VALUE]"
      - "OVH_CONSUMER_KEY=[YOUR_OWN_VALUE]"
    ```

- Replace `postmaster@example.com` by your **own email** within the `certificatesresolvers.myresolver.acme.email` command line argument of the `traefik` service.
- Replace `whoami.example.com` by your **own domain** within the `traefik.http.routers.whoami.rule` label of the `whoami` service.
- Optionally uncomment the following lines if you want to test/debug: 

	```yaml
	#- "--log.level=DEBUG"
	#- "--certificatesresolvers.myresolver.acme.caserver=https://acme-staging-v02.api.letsencrypt.org/directory"
	```

- Run `docker-compose up -d` within the folder where you created the previous file.
- Wait a bit and visit `https://your_own_domain` to confirm everything went fine.

!!! Note

    If you uncommented the `acme.caserver` line, you will get an SSL error, but if you display the certificate and see it was emitted by `Fake LE Intermediate X1` then it means all is good.
    (It is the staging environment intermediate certificate used by let's encrypt).
    You can now safely comment the `acme.caserver` line, remove the `letsencrypt/acme.json` file and restart Traefik to issue a valid certificate.

## Explanation

What changed between the initial setup:

- We configure a second entry point for the https traffic:

```yaml
command:
  # Traefik will listen to incoming request on the port 443 (https)
  - "--entrypoints.websecure.address=:443"
ports:
  - "443:443"
```

- We configure the DNS let's encrypt challenge:

```yaml
command:
  # Enable a dns challenge named "myresolver"
  - "--certificatesresolvers.myresolver.acme.dnschallenge=true"
  # Tell which provider to use
  - "--certificatesresolvers.myresolver.acme.dnschallenge.provider=ovh"
  # The email to provide to let's encrypt
  - "--certificatesresolvers.myresolver.acme.email=postmaster@example.com"
```

- We provide the required configuration to our provider via environment variables: 

```yaml
environment:
  - "OVH_ENDPOINT=xxx"
  - "OVH_APPLICATION_KEY=xxx"
  - "OVH_APPLICATION_SECRET=xxx"
  - "OVH_CONSUMER_KEY=xxx"
```

!!! Note

    This is the step that may vary depending on the provider you use.
    Just define the variables required by your provider.
    (see the prerequisite for a list)

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
	- "traefik.http.routers.whoami.tls.certresolver=myresolver" # Uses the Host rule to define which certificate to issue
```

## Use Secrets

To configure the provider, and avoid having the secrets exposed in plaintext within the docker-compose environment section,
you could use docker secrets.  
The point is to manage those secret files by another mean, and read them from the `docker-compose.yml` file making the docker-compose file itself less sensitive.

- Create a directory named `secrets`, and create a file for each parameters required to configure you provider containing the value of the parameter:  
	for example, the `ovh_endpoint.secret` file contain `ovh-eu`

```text
./secrets
├── ovh_application_key.secret
├── ovh_application_secret.secret
├── ovh_consumer_key.secret
└── ovh_endpoint.secret
```

!!! Note

    You could store those secrets anywhere on the server,
    just make sure to use the proper path for the `file` directive for the secrets definition in the `docker-compose.yml` file.

- Use this `docker-compose.yml` file:

```yaml
--8<-- "content/user-guides/docker-compose/acme-dns/docker-compose_secrets.yml"
```

!!! Note

    Still think about changing `postmaster@example.com` & `whoami.example.com` by your own values.

Let's explain a bit what we just did:

- The following section allow to read files on the docker host, and expose those file under `/run/secrets/[NAME_OF_THE_SECRET]` within the container:

```yaml
secrets:
  # secret name also used to name the file exposed within the container
  ovh_endpoint:
     # path on the host
    file: "./secrets/ovh_endpoint.secret"
  ovh_application_key:
    file: "./secrets/ovh_application_key.secret"
  ovh_application_secret:
    file: "./secrets/ovh_application_secret.secret"
  ovh_consumer_key:
    file: "./secrets/ovh_consumer_key.secret"

services:
  traefik:
    # expose the predefined secret to the container by name
    secrets:
      - "ovh_endpoint"
      - "ovh_application_key"
      - "ovh_application_secret"
      - "ovh_consumer_key"
```

- The environment variable within our `whoami` service are suffixed by `_FILE` which allow us to point to files containing the value, instead of exposing the value itself.  
	The acme client will read the content of those file to get the required configuration values.

```yaml
environment:
  # expose the path to file provided by docker containing the value we want for OVH_ENDPOINT.
  - "OVH_ENDPOINT_FILE=/run/secrets/ovh_endpoint"
  - "OVH_APPLICATION_KEY_FILE=/run/secrets/ovh_application_key"
  - "OVH_APPLICATION_SECRET_FILE=/run/secrets/ovh_application_secret"
  - "OVH_CONSUMER_KEY_FILE=/run/secrets/ovh_consumer_key"
```

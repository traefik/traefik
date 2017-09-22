# How to generate the self-signed wildcard certificate

```bash
#!/usr/bin/env bash

# Specify where we will install
# the wildcard certificate
SSL_DIR="./ssl"

# Set the wildcarded domain
# we want to use
DOMAIN="*.acme.wtf"

# A blank passphrase
PASSPHRASE=""

# Set our CSR variables
SUBJ="
C=FR
ST=MP
O=
localityName=Toulouse
commonName=$DOMAIN
organizationalUnitName=Traefik
emailAddress=
"

# Create our SSL directory
# in case it doesn't exist
sudo mkdir -p "$SSL_DIR"

# Generate our Private Key, CSR and Certificate
sudo openssl genrsa -out "$SSL_DIR/wildcard.key" 2048
sudo openssl req -new -subj "$(echo -n "$SUBJ" | tr "\n" "/")" -key "$SSL_DIR/wildcard.key" -out "$SSL_DIR/wildcard.csr" -passin pass:$PASSPHRASE
sudo openssl x509 -req -days 3650 -in "$SSL_DIR/wildcard.csr" -signkey "$SSL_DIR/wildcard.key" -out "$SSL_DIR/wildcard.crt"
sudo rm -f "$SSL_DIR/wildcard.csr"
```

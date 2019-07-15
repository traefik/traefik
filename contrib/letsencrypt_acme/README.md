## Let's Encrypt Acme Import/Export

The two scripts provided here are designed to import / export basic certificates issued by Let's Encrypt.

Two scenarios are addressed by this:

 - Using a pre-existing Let's Encrypt certificate with Traefik ( such as when migrating during a DNS change to Traefik when previously using Nginx )
 - Migrating away from Traefik to Nginx

### Import - Migrating to Traefik

#### View usage

    ./import.pl --help

#### Import My Site

    ./import.pl --file acme_part.json --cert mySite.pem --key mySiteKey.pem --chain chain.pem

In this case chain.pem is the intermediate certificate of Let's Encrypt.

The acme_part.json file is generated, and the "Certificates" portion of it should be copied and pasted into your Traefik acme.json file. For safety and simplicity the script is not designed to directly modify your acme.json

### Export - Migrating away from Traefik
#### View Usage

    ./export.pl --help

#### List sites in acme.json

    ./export --file /path/to/acme.json

#### Export My Site

    ./export --file /path/to/acme.json --certnum 0

This saves out the following 3 files:

 - cert.pem ( the certificate for your site )
 - privkey.pem ( the private key for your site )
 - chain.pem ( the intermed cert of Let's encrypt )

These files could then be used to configure a Nginx instance, for example, with the certificate last used in Traefik.


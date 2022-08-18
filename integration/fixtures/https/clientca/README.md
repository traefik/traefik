# This is how the certs were created

Password: traefik

```bash
# ca1.example.com
# Country Name (2 letter code) [AU]:.
# State or Province Name (full name) [Some-State]:.
# Locality Name (eg, city) []:.
# Organization Name (eg, company) [Internet Widgits Pty Ltd]:.
# Organizational Unit Name (eg, section) []:.
# Common Name (e.g. server FQDN or YOUR name) []:ca1.example.com
# Email Address []:.
openssl req -new -newkey rsa:2048 -x509 -days 3650 -extensions v3_ca -keyout ca1.pem -out ca1.crt

# ca2.example.com
# Country Name (2 letter code) [AU]:.
# State or Province Name (full name) [Some-State]:.
# Locality Name (eg, city) []:.
# Organization Name (eg, company) [Internet Widgits Pty Ltd]:.
# Organizational Unit Name (eg, section) []:.
# Common Name (e.g. server FQDN or YOUR name) []:ca2.example.com
# Email Address []:.
openssl req -new -newkey rsa:2048 -x509 -days 3650 -extensions v3_ca -keyout ca2.pem -out ca2.crt

# ca3.example.com
# Country Name (2 letter code) [AU]:.
# State or Province Name (full name) [Some-State]:.
# Locality Name (eg, city) []:.
# Organization Name (eg, company) [Internet Widgits Pty Ltd]:.
# Organizational Unit Name (eg, section) []:.
# Common Name (e.g. server FQDN or YOUR name) []:ca3.example.com
# Email Address []:.
openssl req -new -newkey rsa:2048 -x509 -days 3650 -extensions v3_ca -keyout ca3.pem -out ca3.crt

openssl rsa -in ca1.pem -out ca1.key
openssl rsa -in ca2.pem -out ca2.key
openssl rsa -in ca3.pem -out ca3.key
cat ca1.crt ca2.crt > ca1and2.crt
rm ca1.pem ca2.pem ca3.pem

openssl genrsa -out client1.key 2048
openssl genrsa -out client2.key 2048
openssl genrsa -out client3.key 2048

# Country Name (2 letter code) [AU]:.
# State or Province Name (full name) [Some-State]:.
# Locality Name (eg, city) []:.
# Organization Name (eg, company) [Internet Widgits Pty Ltd]:.
# Organizational Unit Name (eg, section) []:.
# Common Name (e.g. server FQDN or YOUR name) []:clien1.example.com
# Email Address []:.
# 
# Please enter the following 'extra' attributes
# to be sent with your certificate request
# A challenge password []:
#
# Issuer
#   CN = ca1.example.com
# Subject
#   CN = clien1.example.com
openssl req -key client1.key -new -out client1.csr

# Country Name (2 letter code) [AU]:.
# State or Province Name (full name) [Some-State]:.
# Locality Name (eg, city) []:.
# Organization Name (eg, company) [Internet Widgits Pty Ltd]:.
# Organizational Unit Name (eg, section) []:.
# Common Name (e.g. server FQDN or YOUR name) []:client2.example.com
# Email Address []:.
# 
# Please enter the following 'extra' attributes
# to be sent with your certificate request
# A challenge password []:
#
# Issuer
#   CN = ca2.example.com
# Subject
#   CN = client2.example.com
openssl req -key client2.key -new -out client2.csr

# Country Name (2 letter code) [AU]:.
# State or Province Name (full name) [Some-State]:.
# Locality Name (eg, city) []:.
# Organization Name (eg, company) [Internet Widgits Pty Ltd]:.
# Organizational Unit Name (eg, section) []:.
# Common Name (e.g. server FQDN or YOUR name) []:client3.example.com
# Email Address []:.
# 
# Please enter the following 'extra' attributes
# to be sent with your certificate request
# A challenge password []:
#
# Issuer
#   CN = ca3.example.com
# Subject
#   CN = client3.example.com
openssl req -key client3.key -new -out client3.csr

openssl x509 -req -days 3650 -in client1.csr -CA ca1.crt -CAkey ca1.key -CAcreateserial -out client1.crt
openssl x509 -req -days 3650 -in client2.csr -CA ca2.crt -CAkey ca2.key -CAcreateserial -out client2.crt
openssl x509 -req -days 3650 -in client3.csr -CA ca3.crt -CAkey ca3.key -CAcreateserial -out client3.crt

```

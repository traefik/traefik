```bash
openssl req -new -newkey rsa:2048 -x509 -days 3650 -nodes -extensions v3_ca -keyout root.key -out root.pem
openssl genrsa -out server.key 2048
openssl req -nodes -key server.key -new -out server.csr
openssl x509 -req -days 3650 -in server.csr -CA root.pem -CAkey root.key -CAcreateserial -out server.pem
```

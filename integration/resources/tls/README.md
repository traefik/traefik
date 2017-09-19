# TLS certificate description

## local.crt / local.key

Generate with
```bash
go run $GOROOT/src/crypto/tls/generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,localhost --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
mv cert.pem local.cert
mv key.pem local.key
```

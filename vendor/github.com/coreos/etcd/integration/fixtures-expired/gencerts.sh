#!/usr/bin/env bash
set -e

if ! [[ "$0" =~ "./gencerts.sh" ]]; then
	echo "must be run from 'fixtures-expired'"
	exit 255
fi

if which cfssl >/dev/null; then
    echo "cfssl is installed; generating certs"
else
    echo "cfssl is not installed; exiting"
    exit 255
fi

cat > ./etcd-root-ca-csr.json <<EOF
{
  "key": {
    "algo": "rsa",
    "size": 4096
  },
  "names": [
    {
      "O": "etcd",
      "OU": "etcd Security",
      "L": "San Francisco",
      "ST": "California",
      "C": "USA"
    }
  ],
  "CN": "etcd-root-ca",
  "ca": {
    "expiry": "1h"
  }
}
EOF

cfssl gencert --initca=true ./etcd-root-ca-csr.json | cfssljson --bare ./etcd-root-ca

cat > ./etcd-gencert.json <<EOF
{
  "signing": {
    "default": {
        "usages": [
          "signing",
          "key encipherment",
          "server auth",
          "client auth"
        ],
        "expiry": "1h"
    }
  }
}
EOF

cat > ./server-ca-csr.json <<EOF
{
  "key": {
    "algo": "rsa",
    "size": 4096
  },
  "names": [
    {
      "O": "etcd",
      "OU": "etcd Security",
      "L": "San Francisco",
      "ST": "California",
      "C": "USA"
    }
  ],
  "CN": "example.com",
  "hosts": [
    "127.0.0.1",
    "localhost"
  ]
}
EOF

cfssl gencert \
    --ca ./etcd-root-ca.pem \
    --ca-key ./etcd-root-ca-key.pem \
    --config ./etcd-gencert.json \
    ./server-ca-csr.json | cfssljson --bare ./server

rm ./*.json
rm ./*.csr

if which openssl >/dev/null; then
    openssl x509 -in ./etcd-root-ca.pem -text -noout
    openssl x509 -in ./server.pem -text -noout
fi

openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -days 3650 \
  -out snitest.com.cert \
  -keyout snitest.com.key \
  -subj "/CN=snitest.com" \
  -addext "subjectAltName = DNS:snitest.com"

openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -days 3650 \
  -out www.snitest.com.cert \
  -keyout www.snitest.com.key \
  -subj "/CN=www.snitest.com" \
  -addext "subjectAltName = DNS:www.snitest.com"

openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -days 3650 \
  -out snitest.org.cert \
  -keyout snitest.org.key \
  -subj "/CN=snitest.org" \
  -addext "subjectAltName = DNS:snitest.org"

openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -days 3650 \
  -out  uppercase_wildcard.www.snitest.com.cert \
  -keyout uppercase_wildcard.www.snitest.com.key \
  -subj "/CN=FOO.WWW.SNITEST.COM" \
  -addext "subjectAltName = DNS:*.WWW.SNITEST.COM"

openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -days 3650 \
  -out  wildcard.www.snitest.com.cert \
  -keyout wildcard.www.snitest.com.key \
  -subj "/CN=*.www.snitest.com" \
  -addext "subjectAltName = DNS:*.www.snitest.com"

openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -days 3650 \
  -out  wildcard.snitest.com.cert \
  -keyout wildcard.snitest.com.key \
  -subj "/CN=*.snitest.com" \
  -addext "subjectAltName = DNS:*.snitest.com"


# ML-DSA-44 fixtures require OpenSSL 3.5 or newer.
mkdir -p mldsa
(
  cd mldsa || exit 1
  for i in 1 2 3; do
    openssl req -newkey ML-DSA-44 -noenc -x509 -days 36500 \
      -keyout "ca${i}.key" -out "ca${i}.crt" -subj "/CN=ca${i}.example.com" \
      -addext "basicConstraints=critical,CA:TRUE" \
      -addext "keyUsage=critical,keyCertSign,cRLSign"
    openssl req -newkey ML-DSA-44 -noenc -new \
      -keyout "client${i}.key" -out "client${i}.csr" -subj "/CN=client${i}"
    printf 'extendedKeyUsage=clientAuth\n' > extensions
    openssl x509 -req -in "client${i}.csr" -CA "ca${i}.crt" -CAkey "ca${i}.key" \
      -set_serial "$i" -days 36500 -extfile extensions -out "client${i}.crt"
    openssl pkey -provparam ml-dsa.output_formats=seed-only \
      -in "client${i}.key" -out "client${i}-seed.key"
    mv "client${i}-seed.key" "client${i}.key"
  done
  for host in snitest.com snitest.org; do
    openssl req -newkey ML-DSA-44 -noenc -x509 -days 36500 \
      -keyout "$host.key" -out "$host.cert" -subj "/CN=$host" \
      -addext "subjectAltName=DNS:$host" -addext "extendedKeyUsage=serverAuth"
    openssl pkey -provparam ml-dsa.output_formats=seed-only \
      -in "$host.key" -out "$host-seed.key"
    mv "$host-seed.key" "$host.key"
  done
  rm ca*.key ca3.crt client*.csr extensions
)

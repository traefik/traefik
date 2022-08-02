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

openssl req -newkey rsa:2048 \
  -new -nodes -x509 \
  -days 3650 \
  -out alt.snitest.com_www.snitest.com.cert \
  -keyout alt.snitest.com_www.snitest.com.key \
  -subj "/CN=www.snitest.com" \
  -addext "subjectAltName = DNS:www.snitest.com, DNS:alt.snitest.com"

openssl ecparam -name prime256v1 -genkey -noout -out ecdsa.snitest.com_www.snitest.com.key
openssl req -new -nodes -x509 -days 3650 -sha256 -key ecdsa.snitest.com_www.snitest.com.key -out ecdsa.snitest.com_www.snitest.com.cert -subj "/CN=www.snitest.com" -addext "subjectAltName = DNS:www.snitest.com, DNS:ecdsa.snitest.com"
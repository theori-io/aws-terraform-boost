#!/usr/bin/env bash
set -e
CERT_PATH=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)/certs

case `uname -s` in
    Linux*)     sslConfig=/etc/ssl/openssl.cnf;;
    Darwin*)    sslConfig=/System/Library/OpenSSL/openssl.cnf;;
esac

mkdir -p $CERT_PATH
cd $CERT_PATH

openssl genrsa -out ca.key 2048
openssl req -key rsa:2048 -x509 -new -nodes -key ca.key -subj /C=US/O=Amazon/CN='Amazon Fake' -sha256 -days 1825 -out ca.pem \
    -reqexts v3_ca \
    -extensions v3_ca \
    -config <(cat $sslConfig \
        <(printf '[v3_ca]\nkeyUsage=keyCertSign')) \

openssl req -new \
    -newkey rsa:2048 \
    -nodes \
    -keyout server.key \
    -out server.key.csr \
    -subj /CN=*.amazonaws.com

openssl x509 -req \
    -in server.key.csr \
    -CA ca.pem \
    -CAkey ca.key \
    -CAcreateserial \
    -out server.pem \
    -sha256 \
    -days 365 \
    -extfile - << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
subjectAltName = DNS:*.af-south-1.amazonaws.com,DNS:*.ap-east-1.amazonaws.com,DNS:*.ap-northeast-1.amazonaws.com,DNS:*.ap-northeast-2.amazonaws.com,DNS:*.ap-northeast-3.amazonaws.com,DNS:*.ap-south-1.amazonaws.com,DNS:*.ap-south-2.amazonaws.com,DNS:*.ap-southeast-1.amazonaws.com,DNS:*.ap-southeast-2.amazonaws.com,DNS:*.ap-southeast-3.amazonaws.com,DNS:*.ap-southeast-4.amazonaws.com,DNS:*.ca-central-1.amazonaws.com,DNS:*.eu-central-1.amazonaws.com,DNS:*.eu-central-2.amazonaws.com,DNS:*.eu-north-1.amazonaws.com,DNS:*.eu-south-1.amazonaws.com,DNS:*.eu-south-2.amazonaws.com,DNS:*.eu-west-1.amazonaws.com,DNS:*.eu-west-2.amazonaws.com,DNS:*.eu-west-3.amazonaws.com,DNS:*.il-central-1.amazonaws.com,DNS:*.me-central-1.amazonaws.com,DNS:*.me-south-1.amazonaws.com,DNS:*.sa-east-1.amazonaws.com,DNS:*.us-east-1.amazonaws.com,DNS:*.us-east-2.amazonaws.com,DNS:*.us-west-1.amazonaws.com,DNS:*.us-west-2.amazonaws.com
EOF

cat server.pem ca.pem > fullchain.pem

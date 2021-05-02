#!/bin/bash
# This script generates test & development certificates. Not for production use!

O="IRIS"
ST="Berlin"
L="Berlin"
C="DE"
OU="IT"
CN="Testing-Development"
# using less than 1024 here will result in a TLS handshake failure in Go
# using less than 2048 will cause e.g. 'curl' to complain that the ciper is too weak
LEN="2048"


openssl genrsa -out rootCA.key ${LEN}
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.crt -subj "/C=${C}/ST=${ST}/L=${L}/O=${O}/OU=${OU}/CN=${CN}"

declare -a certs=("grpc-client" "grpc-server" "jsonrpc-client" "jsonrpc-server")

for cert in "${certs[@]}"
do
	echo "Generating and signing certificates for ${cert}...";
	openssl genrsa -out "${cert}.key" ${LEN};
	openssl rsa -in "${cert}.key" -pubout -out "${cert}.pub";
	openssl req -new -sha256 -key "${cert}.key" -subj "/C=${C}/ST=${ST}/L=${L}/O=${O}/OU=${OU}/CN=${cert}" -out "${cert}.csr";
	openssl x509 -req -in "${cert}.csr" -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out "${cert}.crt" -days 500 -sha256;
done
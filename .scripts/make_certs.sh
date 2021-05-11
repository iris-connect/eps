#!/bin/bash
# This script generates test & development certificates. Not for production use!

# You can addd an entry to this list to generate a certificate for a given
# operator. The name of the operator will be added as the common name as well
# as the subject alternative name (SAN), which is required for some newer
# TLS libraries.
declare -a certs=("op-1" "op-2" "hd-1" "ls-1" "internal-server" "public-proxy-1" "private-proxy-1" "sd-1")

O="IRIS"
ST="Berlin"
L="Berlin"
C="DE"
OU="IT"
CN="Testing-Development"
# using less than 1024 here will result in a TLS handshake failure in Go
# using less than 2048 will cause e.g. 'curl' to complain that the ciper is too weak
LEN="2048"

# Please note that we add a ".local" name to the wildcard subject alternative names as
# second-level wildcards (e.g. "*.internal-server") will not work. Probably a good security
# measure as one could otherwise register a wildcard like "*.com"

openssl genrsa -out root.key ${LEN}
openssl req -x509 -new -nodes -key root.key -sha256 -days 1024 -out root.crt -subj "/C=${C}/ST=${ST}/L=${L}/O=${O}/OU=${OU}/CN=${CN}"

for cert in "${certs[@]}"
do
	echo "Generating and signing certificates for ${cert}...";
	openssl genrsa -out "${cert}.key" ${LEN};
	openssl rsa -in "${cert}.key" -pubout -out "${cert}.pub";
	openssl req -new -sha256 -key "${cert}.key" -subj "/C=${C}/ST=${ST}/L=${L}/O=${O}/OU=${OU}/CN=${cert}" -addext "subjectAltName = DNS:${cert},DNS:*.${cert}.local" -out "${cert}.csr";
	openssl x509 -req -in "${cert}.csr" -CA root.crt -CAkey root.key -CAcreateserial -out "${cert}.crt" -extensions SAN -extfile <(printf "[SAN]\nsubjectAltName = DNS:${cert},DNS:*.${cert}.local") -days 500 -sha256;

	# the signing certificates use ECDSA and are for signing service directory entries
	openssl ecparam -genkey -name prime256v1 -noout -out "${cert}-sign.key";
	openssl ec -in "${cert}-sign.key" -pubout -out "${cert}-sign.pub";
	openssl req -new -sha256 -key "${cert}-sign.key" -subj "/C=${C}/ST=${ST}/L=${L}/O=${O}/OU=${OU}/CN=${cert}" -addext "keyUsage=digitalSignature" -addext "subjectAltName = DNS:${cert},DNS:*.${cert}.local"  -out "${cert}-sign.csr";
	openssl x509 -req -in "${cert}-sign.csr" -CA root.crt -CAkey root.key -CAcreateserial -out "${cert}-sign.crt"  -extensions SANKey -extfile <(printf "[SANKey]\nsubjectAltName = DNS:${cert},DNS:*.${cert}.local\nkeyUsage = digitalSignature") -days 500 -sha256;

done
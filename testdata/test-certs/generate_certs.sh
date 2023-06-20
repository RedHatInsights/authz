#!/bin/bash

# Create the client-ca private key and self-signed certificate
openssl req -new -newkey rsa:2048 -nodes -keyout client-ca.key -x509 -days 36500 -out client-ca.crt -subj "/CN=client-ca"

# Create the client private key and certificate signing request (CSR)
openssl req -new -newkey rsa:2048 -nodes -keyout client.key -out client.csr -subj "/CN=authz"

# Sign the client CSR with the client-ca certificate and key
openssl x509 -req -in client.csr -CA client-ca.crt -CAkey client-ca.key -CAcreateserial -out client.crt -days 36500

# Create the server-ca private key and self-signed certificate
openssl req -new -newkey rsa:2048 -nodes -keyout server-ca.key -x509 -days 36500 -out server-ca.crt -subj "/CN=server-ca"

# Create the server private key and certificate signing request (CSR)
openssl req -new -newkey rsa:2048 -nodes -keyout server.key -out server.csr -subj "/CN=localhost"

# Add subject alternative name (SAN) extension to the server CSR
echo "subjectAltName = IP:127.0.0.1" > extfile.cnf

# Sign the server CSR with the server-ca certificate and key
openssl x509 -req -in server.csr -CA server-ca.crt -CAkey server-ca.key -CAcreateserial -out server.crt -days 36500 -extfile extfile.cnf

# Clean up temporary files
rm client.csr server.csr extfile.cnf

echo "Certificates generated successfully!"

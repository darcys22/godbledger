#!/bin/bash

export IP="127.0.0.1"
rm *.pem

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=AU/ST=Victoria/L=Melbourne/O=Darcy Financial/OU=GoDBLedger/CN=*.darcyfinancial.com/emailAddress=sean@darcyfinancial.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=AU/ST=Victoria/L=Melbourne/O=PC Server/OU=Computer/CN=*.pcserver.com/emailAddress=pcserver@gmail.com"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
echo subjectAltName = IP:${IP} > server-ext.cnf
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text

# 4. Generate client's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=AU/ST=Victoria/L=Melbourne/O=PC Client/OU=Computer/CN=*.pcclient.com/emailAddress=pcclient@gmail.com"

# 5. Use CA's private key to sign client's CSR and get back the signed certificate
echo subjectAltName = IP:${IP} > client-ext.cnf
openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf
echo "Client's signed certificate"
openssl x509 -in client-cert.pem -noout -text

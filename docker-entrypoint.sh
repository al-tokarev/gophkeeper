#!/bin/sh
set -e

CERT_DIR="${TLS_CERT_DIR:-/certs}"
CA_CERT="$CERT_DIR/ca.crt"
CA_KEY="$CERT_DIR/ca.key"
SRV_CERT="$CERT_DIR/server.crt"
SRV_KEY="$CERT_DIR/server.key"

if [ ! -f "$CA_CERT" ] || [ ! -f "$SRV_CERT" ] || [ ! -f "$SRV_KEY" ]; then
    echo "Generating CA and server certificate..."
    mkdir -p "$CERT_DIR"

    openssl req -x509 -newkey rsa:4096 \
        -keyout "$CA_KEY" -out "$CA_CERT" \
        -days 365 -nodes \
        -subj "/CN=GophKeeper Local CA" \
        -addext "basicConstraints=critical,CA:TRUE"

    openssl req -newkey rsa:4096 \
        -keyout "$SRV_KEY" -out "$CERT_DIR/server.csr" \
        -nodes \
        -subj "/CN=localhost"

    openssl x509 -req -in "$CERT_DIR/server.csr" \
        -CA "$CA_CERT" -CAkey "$CA_KEY" \
        -CAcreateserial -out "$SRV_CERT" \
        -days 365 \
        -extfile <(printf "subjectAltName=DNS:localhost,DNS:server,IP:127.0.0.1\nbasicConstraints=CA:FALSE")

    chmod 600 "$CA_KEY" "$SRV_KEY"
    chmod 644 "$CA_CERT" "$SRV_CERT"
    echo "Certificates generated: $CA_CERT, $SRV_CERT"
else
    echo "Certificates already exist"
fi

exec /app/server
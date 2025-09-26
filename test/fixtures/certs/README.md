# Test Certificate Fixtures

**WARNING: These are test certificates only - DO NOT use in production!**

This directory contains self-signed certificates for testing TLS functionality.

## Files

- `ca.crt` - Certificate Authority (CA) certificate for test purposes
- `client.crt` - Client certificate signed by the test CA
- `client.key` - Private key for the client certificate

## Certificate Details

**CA Certificate:**
- Subject: C=US, ST=CA, O=Test, CN=TestCA
- Valid for: 10 years
- Self-signed

**Client Certificate:**
- Subject: C=US, ST=CA, O=Test, CN=TestClient
- Valid for: 1 year
- Signed by TestCA

## Regeneration Commands

If you need to regenerate these certificates:

```bash
# Generate CA private key
openssl genrsa -out ca.key 2048

# Generate CA certificate
openssl req -new -x509 -key ca.key -sha256 -subj "/C=US/ST=CA/O=Test/CN=TestCA" -days 3650 -out ca.crt

# Generate client private key
openssl genrsa -out client.key 2048

# Generate client certificate request
openssl req -new -key client.key -subj "/C=US/ST=CA/O=Test/CN=TestClient" -out client.csr

# Sign client certificate with CA
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256

# Clean up temporary files
rm -f ca.key client.csr ca.srl
```

## Usage in Tests

These certificates can be used to test:
- TLS client authentication
- Certificate validation
- Insecure skip verify functionality
- CA certificate trust chains
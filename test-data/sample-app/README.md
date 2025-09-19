# Sample Application for Image Builder Demo

This is a simple demo application used to test the OCI image building functionality.

## Files

- `app.py` - Simple Python web server
- `Dockerfile` - Container build instructions
- `requirements.txt` - Python dependencies
- `config.json` - Application configuration

## Usage

When built as an OCI image, this application:

1. Starts a web server on port 8080
2. Serves a demo page at `/`
3. Provides health check at `/health`
4. Shows build timestamp and environment

## Build Example

```bash
./demo-features.sh build-image \
  --context ./test-data/sample-app \
  --tag myapp:v1.0 \
  --storage /tmp/oci-storage
```

This demonstrates the complete OCI image building process using the image builder package.
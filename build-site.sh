#!/bin/bash
set -e

# Build script for generating the static site
# This script prepares the site for deployment to Cloudflare Pages

echo "Building IDP Builder static site..."

# Create output directory
BUILD_DIR="${BUILD_DIR:-./site}"
OUTPUT_DIR="${OUTPUT_DIR:-./public}"

echo "Source directory: $BUILD_DIR"
echo "Output directory: $OUTPUT_DIR"

# Clean output directory
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Copy all static files
echo "Copying static files..."
cp -r "$BUILD_DIR"/* "$OUTPUT_DIR/"

# Create _headers file for Cloudflare Pages (optional security headers)
echo "Creating _headers file..."
cat > "$OUTPUT_DIR/_headers" << 'EOF'
/*
  X-Frame-Options: DENY
  X-Content-Type-Options: nosniff
  X-XSS-Protection: 1; mode=block
  Referrer-Policy: strict-origin-when-cross-origin
  Permissions-Policy: accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()

/*.css
  Cache-Control: public, max-age=31536000, immutable

/*.js
  Cache-Control: public, max-age=31536000, immutable

/index.html
  Cache-Control: public, max-age=0, must-revalidate
EOF

# Create _redirects file for Cloudflare Pages (optional redirects)
echo "Creating _redirects file..."
cat > "$OUTPUT_DIR/_redirects" << 'EOF'
# Redirect /docs to /docs/index.html
/docs /docs/index.html 200

# 404 page
/* /404.html 404
EOF

echo "Build completed successfully!"
echo "Site built to: $OUTPUT_DIR"

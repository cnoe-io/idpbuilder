#!/bin/bash
set -e

# Build script for generating the static site
# This script prepares the site for deployment to Cloudflare Pages

echo "Building IDP Builder static site..."

# Install npm dependencies if needed
if [ -f "package.json" ]; then
    echo "Checking npm dependencies..."
    if command -v npm >/dev/null 2>&1; then
        # Only install if node_modules doesn't exist or is missing dependencies
        if [ ! -d "node_modules" ] || [ ! -d "node_modules/marked" ]; then
            echo "Installing npm dependencies..."
            npm install --quiet
        fi
    else
        echo "Warning: npm not found. Skipping dependency installation."
    fi
fi

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

# Copy organized documentation from docs/ to public/docs/
echo "Copying organized documentation..."
DOCS_SOURCE_DIR="${DOCS_SOURCE_DIR:-./docs}"
if [ -d "$DOCS_SOURCE_DIR" ]; then
    # Create docs directory structure in output
    mkdir -p "$OUTPUT_DIR/docs/specs"
    mkdir -p "$OUTPUT_DIR/docs/implementation"
    mkdir -p "$OUTPUT_DIR/docs/user"
    mkdir -p "$OUTPUT_DIR/docs/images"
    
    # Copy markdown files (will be converted to HTML next)
    cp -r "$DOCS_SOURCE_DIR/specs"/*.md "$OUTPUT_DIR/docs/specs/" 2>/dev/null || true
    cp -r "$DOCS_SOURCE_DIR/implementation"/*.md "$OUTPUT_DIR/docs/implementation/" 2>/dev/null || true
    cp -r "$DOCS_SOURCE_DIR/user"/*.md "$OUTPUT_DIR/docs/user/" 2>/dev/null || true
    cp -r "$DOCS_SOURCE_DIR/images"/* "$OUTPUT_DIR/docs/images/" 2>/dev/null || true
    
    # Copy main docs README if it exists
    [ -f "$DOCS_SOURCE_DIR/README.md" ] && cp "$DOCS_SOURCE_DIR/README.md" "$OUTPUT_DIR/docs/README.md"
    
    echo "Documentation copied successfully!"
    
    # Convert markdown to HTML
    echo "Converting markdown documentation to HTML..."
    if command -v node >/dev/null 2>&1; then
        if [ -f "./scripts/convert-markdown.js" ]; then
            DOCS_SOURCE_DIR="$DOCS_SOURCE_DIR" OUTPUT_DIR="$OUTPUT_DIR" node ./scripts/convert-markdown.js
            
            # Remove the markdown source files after conversion
            find "$OUTPUT_DIR/docs/specs" -name "*.md" -type f ! -name "README.md" -delete 2>/dev/null || true
            find "$OUTPUT_DIR/docs/implementation" -name "*.md" -type f ! -name "README.md" -delete 2>/dev/null || true
            find "$OUTPUT_DIR/docs/user" -name "*.md" -type f ! -name "README.md" -delete 2>/dev/null || true
            
            echo "Markdown conversion completed!"
        else
            echo "Warning: Conversion script not found. Markdown files will be served as-is."
        fi
    else
        echo "Warning: Node.js not found. Markdown files will be served as-is."
    fi
else
    echo "Warning: Documentation source directory not found at $DOCS_SOURCE_DIR"
fi

# Copy examples from examples/ to public/docs/examples/
echo "Copying examples..."
EXAMPLES_SOURCE_DIR="${EXAMPLES_SOURCE_DIR:-./examples}"
if [ -d "$EXAMPLES_SOURCE_DIR" ]; then
    # Create examples directory in output
    mkdir -p "$OUTPUT_DIR/docs/examples"
    mkdir -p "$OUTPUT_DIR/docs/examples/v1alpha2"
    
    # Copy all example files and directories
    cp -r "$EXAMPLES_SOURCE_DIR"/* "$OUTPUT_DIR/docs/examples/" 2>/dev/null || true
    
    echo "Examples copied successfully!"
    
    # Convert example README markdown files to HTML
    if command -v node >/dev/null 2>&1; then
        if [ -f "./scripts/convert-examples.js" ]; then
            OUTPUT_DIR="$OUTPUT_DIR" node ./scripts/convert-examples.js
            
            # Remove the markdown source files after conversion
            find "$OUTPUT_DIR/docs/examples" -name "README.md" -type f -delete 2>/dev/null || true
            
            echo "Examples markdown conversion completed!"
        fi
    fi
else
    echo "Warning: Examples source directory not found at $EXAMPLES_SOURCE_DIR"
fi

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

# Serve markdown files as plain text or allow direct access
/docs/specs/* /docs/specs/:splat 200
/docs/implementation/* /docs/implementation/:splat 200
/docs/user/* /docs/user/:splat 200
/docs/images/* /docs/images/:splat 200
/docs/examples/* /docs/examples/:splat 200
EOF

echo "Build completed successfully!"
echo "Site built to: $OUTPUT_DIR"

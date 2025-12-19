# IDP Builder Static Site

This directory contains the static website for IDP Builder, configured for deployment on Cloudflare Pages.

## Directory Structure

```
site/
├── index.html          # Homepage
├── 404.html           # Custom 404 page
├── css/
│   └── style.css      # Main stylesheet
├── js/               # JavaScript files (if needed)
└── docs/
    └── index.html     # Documentation page
```

## Local Development

### Prerequisites

- Node.js and npm (for development server)
- Bash (for build script)

### Running Locally

1. Install dependencies:
   ```bash
   npm install
   ```

2. Start development server:
   ```bash
   npm run dev
   ```
   This will serve the site from `./site` directory on http://localhost:8080

3. To preview the built site:
   ```bash
   npm run build
   npm run preview
   ```
   This will build and serve from `./public` directory

## Building for Production

Run the build script to prepare the site for deployment:

```bash
./build-site.sh
```

This will:
- Copy all files from `site/` to `public/`
- Create `_headers` file with security headers
- Create `_redirects` file for routing
- Optimize for Cloudflare Pages deployment

## Deployment to Cloudflare Pages

### Option 1: Automatic Deployment (Recommended)

1. **Connect Repository to Cloudflare Pages:**
   - Go to [Cloudflare Dashboard](https://dash.cloudflare.com/)
   - Navigate to Pages
   - Click "Create a project"
   - Connect your GitHub repository
   - Select the repository: `greghaynes/idpbuilder`

2. **Configure Build Settings:**
   - **Build command:** `./build-site.sh`
   - **Build output directory:** `public`
   - **Root directory:** `/` (leave empty or root)
   - **Environment variables:** None required

3. **Deploy:**
   - Click "Save and Deploy"
   - Cloudflare will automatically build and deploy on every push to the main branch

### Option 2: Manual Deployment with Wrangler CLI

1. **Install Wrangler:**
   ```bash
   npm install -g wrangler
   ```

2. **Login to Cloudflare:**
   ```bash
   wrangler login
   ```

3. **Deploy:**
   ```bash
   npm run build
   wrangler pages deploy public --project-name=idpbuilder
   ```

### Option 3: Direct Upload

1. Build the site:
   ```bash
   npm run build
   ```

2. Go to Cloudflare Pages dashboard
3. Click "Upload assets"
4. Upload the contents of the `public` directory

## Build Configuration

### Environment Variables

You can customize the build by setting environment variables:

- `BUILD_DIR`: Source directory (default: `./site`)
- `OUTPUT_DIR`: Output directory (default: `./public`)

Example:
```bash
BUILD_DIR=./site OUTPUT_DIR=./dist ./build-site.sh
```

### Custom Headers

The build script creates a `_headers` file with security headers:
- X-Frame-Options
- X-Content-Type-Options
- X-XSS-Protection
- Referrer-Policy
- Cache-Control for static assets

### Redirects

The `_redirects` file handles:
- `/docs` → `/docs/index.html`
- Custom 404 page for all unmatched routes

## Customization

### Adding New Pages

1. Create HTML file in `site/` directory
2. Link to `/css/style.css` for consistent styling
3. Rebuild the site with `./build-site.sh`

### Modifying Styles

Edit `site/css/style.css` and rebuild.

### Adding JavaScript

1. Create JS files in `site/js/`
2. Reference in your HTML files
3. Rebuild the site

## Performance Optimization

The site is optimized for Cloudflare Pages with:
- Minified and cached CSS
- Proper cache headers for static assets
- Security headers for enhanced protection
- Fast global CDN delivery via Cloudflare

## Troubleshooting

### Build fails on Cloudflare Pages

- Ensure `build-site.sh` has execute permissions: `chmod +x build-site.sh`
- Check that all referenced files exist in the `site/` directory

### Styles not loading

- Verify CSS file exists at `site/css/style.css`
- Check browser console for 404 errors
- Ensure paths start with `/` for absolute references

### 404 page not showing

- Verify `404.html` exists in the build output
- Check `_redirects` file was created correctly

## Additional Resources

- [Cloudflare Pages Documentation](https://developers.cloudflare.com/pages/)
- [Cloudflare Pages Deployment Guide](https://developers.cloudflare.com/pages/get-started/)
- [Wrangler CLI Documentation](https://developers.cloudflare.com/workers/wrangler/)

## License

Apache License 2.0 - See LICENSE file for details

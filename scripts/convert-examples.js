#!/usr/bin/env node

/**
 * Convert example README markdown files to HTML for the static site
 */

const fs = require('fs');
const path = require('path');
const { marked } = require('marked');

marked.setOptions({ gfm: true, breaks: false, headerIds: true, mangle: false });

const createHtmlPage = (title, content, breadcrumbPath = '') => `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="${title} - IDP Builder Examples">
    <title>${title} | IDP Builder</title>
    <link rel="stylesheet" href="/css/style.css">
    <style>
        .docs-container { max-width: 900px; margin: 2rem auto; padding: 0 1rem; }
        .breadcrumb { font-size: 0.9rem; color: var(--text-muted); margin-bottom: 1rem; }
        .breadcrumb a { color: var(--primary-color); text-decoration: none; }
        .breadcrumb a:hover { text-decoration: underline; }
        .markdown-content { line-height: 1.6; }
        .markdown-content h1 { border-bottom: 2px solid var(--border-color); padding-bottom: 0.5rem; margin-bottom: 1.5rem; }
        .markdown-content h2 { margin-top: 2rem; padding-bottom: 0.3rem; border-bottom: 1px solid var(--border-color); }
        .markdown-content h3 { margin-top: 1.5rem; }
        .markdown-content pre { background-color: #f6f8fa; padding: 1rem; border-radius: 6px; overflow-x: auto; }
        .markdown-content code { font-family: 'Courier New', monospace; background-color: #f6f8fa; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.9em; }
        .markdown-content pre code { background-color: transparent; padding: 0; }
        .markdown-content ul, .markdown-content ol { margin-bottom: 1rem; }
        .markdown-content table { border-collapse: collapse; width: 100%; margin: 1rem 0; }
        .markdown-content table th, .markdown-content table td { border: 1px solid var(--border-color); padding: 0.5rem; }
        .markdown-content table th { background-color: var(--bg-alt); }
        .markdown-content blockquote { border-left: 4px solid var(--primary-color); padding-left: 1rem; margin: 1rem 0; color: var(--text-muted); }
    </style>
</head>
<body>
    <header>
        <nav class="container">
            <div class="logo"><h1>IDP Builder</h1></div>
            <button class="menu-toggle" aria-label="Toggle menu" onclick="toggleMenu()">☰</button>
            <ul class="nav-links" id="navLinks">
                <li><a href="/">Home</a></li>
                <li><a href="/docs">Docs</a></li>
                <li><a href="/docs/examples.html">Examples</a></li>
                <li><a href="https://github.com/cnoe-io/idpbuilder" target="_blank" rel="noopener">GitHub</a></li>
            </ul>
        </nav>
    </header>
    <main class="container">
        <div class="docs-container">
            ${breadcrumbPath ? `<div class="breadcrumb">${breadcrumbPath}</div>` : ''}
            <div class="markdown-content">${content}</div>
        </div>
    </main>
    <footer>
        <div class="container">
            <p>&copy; 2024 CNOE IDP Builder. Licensed under <a href="https://github.com/cnoe-io/idpbuilder/blob/main/LICENSE" target="_blank" rel="noopener">Apache License 2.0</a></p>
        </div>
    </footer>
    <script>
        function toggleMenu() {
            const navLinks = document.getElementById('navLinks');
            navLinks.classList.toggle('active');
        }
        document.addEventListener('click', function(event) {
            const nav = document.querySelector('nav');
            const navLinks = document.getElementById('navLinks');
            if (!nav.contains(event.target)) { navLinks.classList.remove('active'); }
        });
    </script>
</body>
</html>`;

const outputDir = process.env.OUTPUT_DIR || './public';

// Convert main examples README
const readmePath = path.join(outputDir, 'docs/examples/README.md');
if (fs.existsSync(readmePath)) {
    const markdown = fs.readFileSync(readmePath, 'utf8');
    const html = marked(markdown);
    const titleMatch = markdown.match(/^#\s+(.+)$/m);
    const title = titleMatch ? titleMatch[1] : 'Examples';
    const breadcrumb = '<a href="/docs">Documentation</a> / <a href="/docs/examples.html">Examples</a> / Platform Examples';
    const fullHtml = createHtmlPage(title, html, breadcrumb);
    const outputPath = path.join(outputDir, 'docs/examples/README.html');
    fs.writeFileSync(outputPath, fullHtml);
    console.log(`Converted: examples/README.md -> ${outputPath}`);
}

// Convert v1alpha2 README
const v1alpha2Path = path.join(outputDir, 'docs/examples/v1alpha2/README.md');
if (fs.existsSync(v1alpha2Path)) {
    const markdown = fs.readFileSync(v1alpha2Path, 'utf8');
    const html = marked(markdown);
    const titleMatch = markdown.match(/^#\s+(.+)$/m);
    const title = titleMatch ? titleMatch[1] : 'V1Alpha2 Examples';
    const breadcrumb = '<a href="/docs">Documentation</a> / <a href="/docs/examples.html">Examples</a> / V1Alpha2 Examples';
    const fullHtml = createHtmlPage(title, html, breadcrumb);
    const outputPath = path.join(outputDir, 'docs/examples/v1alpha2/README.html');
    fs.writeFileSync(outputPath, fullHtml);
    console.log(`Converted: examples/v1alpha2/README.md -> ${outputPath}`);
}

console.log('Examples conversion complete!');

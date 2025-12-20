#!/usr/bin/env node

/**
 * Convert markdown documentation to HTML for the static site
 * This script converts all markdown files in docs/ to HTML pages
 */

const fs = require('fs');
const path = require('path');
const { marked } = require('marked');

// Custom renderer for mermaid diagrams
const renderer = new marked.Renderer();
const originalCodeRenderer = renderer.code.bind(renderer);

renderer.code = function(code, language) {
  if (language === 'mermaid') {
    // Return mermaid code block with the class that mermaid.js will process
    return `<pre class="mermaid">${code}</pre>`;
  }
  return originalCodeRenderer(code, language);
};

// Configure marked for GitHub-flavored markdown
marked.setOptions({
  gfm: true,
  breaks: false,
  headerIds: true,
  mangle: false,
  renderer: renderer
});

// HTML template for documentation pages
const createHtmlPage = (title, content, category, relativePath = '') => {
  const breadcrumb = category ? `<a href="${relativePath}../index.html">${category}</a>` : '';
  
  return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="${title} - IDP Builder Documentation">
    <title>${title} | IDP Builder</title>
    <link rel="stylesheet" href="${relativePath}../../css/style.css">
    <script type="module">
        import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
        mermaid.initialize({ 
            startOnLoad: true,
            theme: 'default',
            securityLevel: 'loose'
        });
    </script>
    <style>
        .docs-container {
            max-width: 900px;
            margin: 2rem auto;
            padding: 0 1rem;
        }
        .breadcrumb {
            font-size: 0.9rem;
            color: var(--text-muted);
            margin-bottom: 1rem;
        }
        .breadcrumb a {
            color: var(--primary-color);
            text-decoration: none;
        }
        .breadcrumb a:hover {
            text-decoration: underline;
        }
        .markdown-content {
            line-height: 1.6;
        }
        .markdown-content h1 {
            border-bottom: 2px solid var(--border-color);
            padding-bottom: 0.5rem;
            margin-bottom: 1.5rem;
        }
        .markdown-content h2 {
            margin-top: 2rem;
            margin-bottom: 1rem;
            border-bottom: 1px solid var(--border-color);
            padding-bottom: 0.3rem;
        }
        .markdown-content h3 {
            margin-top: 1.5rem;
            margin-bottom: 0.75rem;
        }
        .markdown-content pre {
            background-color: var(--bg-alt);
            padding: 1rem;
            border-radius: 5px;
            overflow-x: auto;
        }
        .markdown-content code {
            background-color: var(--bg-alt);
            padding: 0.2rem 0.4rem;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
        }
        .markdown-content pre code {
            background-color: transparent;
            padding: 0;
        }
        .markdown-content blockquote {
            border-left: 4px solid var(--primary-color);
            padding-left: 1rem;
            margin-left: 0;
            color: var(--text-muted);
        }
        .markdown-content table {
            border-collapse: collapse;
            width: 100%;
            margin: 1rem 0;
        }
        .markdown-content th,
        .markdown-content td {
            border: 1px solid var(--border-color);
            padding: 0.5rem;
            text-align: left;
        }
        .markdown-content th {
            background-color: var(--bg-alt);
            font-weight: bold;
        }
        .markdown-content a {
            color: var(--primary-color);
            text-decoration: none;
        }
        .markdown-content a:hover {
            text-decoration: underline;
        }
        .markdown-content img {
            max-width: 100%;
            height: auto;
        }
        .markdown-content ul,
        .markdown-content ol {
            padding-left: 2rem;
            margin: 1rem 0;
        }
        .markdown-content li {
            margin: 0.5rem 0;
        }
        /* Mermaid diagram styling */
        .markdown-content .mermaid {
            background-color: transparent;
            padding: 1rem;
            margin: 1.5rem 0;
            text-align: center;
            overflow-x: auto;
        }
        .markdown-content pre.mermaid {
            background-color: var(--bg-alt);
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <header>
        <nav class="container">
            <div class="logo">
                <h1>IDP Builder</h1>
            </div>
            <button class="menu-toggle" aria-label="Toggle menu" onclick="toggleMenu()">
                ☰
            </button>
            <ul class="nav-links" id="navLinks">
                <li><a href="${relativePath}../../index.html">Home</a></li>
                <li><a href="${relativePath}../index.html">Docs</a></li>
                <li><a href="https://github.com/cnoe-io/idpbuilder" target="_blank" rel="noopener">GitHub</a></li>
            </ul>
        </nav>
    </header>

    <main class="container">
        <div class="docs-container">
            <div class="breadcrumb">
                <a href="${relativePath}../../index.html">Home</a> / 
                <a href="${relativePath}../index.html">Docs</a>${breadcrumb ? ' / ' + breadcrumb : ''}
            </div>
            <article class="markdown-content">
                ${content}
            </article>
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
            
            if (!nav.contains(event.target)) {
                navLinks.classList.remove('active');
            }
        });
    </script>
</body>
</html>`;
};

// Convert a single markdown file to HTML
const convertMarkdownFile = (inputPath, outputPath, category) => {
  try {
    const markdown = fs.readFileSync(inputPath, 'utf8');
    const html = marked(markdown);
    
    // Extract title from first h1 or use filename
    const titleMatch = markdown.match(/^#\s+(.+)$/m);
    const title = titleMatch ? titleMatch[1] : path.basename(inputPath, '.md');
    
    // Calculate relative path based on nesting
    const relativePath = '';
    
    const fullHtml = createHtmlPage(title, html, category, relativePath);
    
    // Create output directory if it doesn't exist
    const outputDir = path.dirname(outputPath);
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }
    
    fs.writeFileSync(outputPath, fullHtml);
    console.log(`Converted: ${inputPath} -> ${outputPath}`);
  } catch (error) {
    console.error(`Error converting ${inputPath}:`, error.message);
  }
};

// Convert all markdown files in a directory
const convertDirectory = (inputDir, outputDir, category) => {
  if (!fs.existsSync(inputDir)) {
    console.log(`Directory not found: ${inputDir}`);
    return;
  }
  
  const files = fs.readdirSync(inputDir);
  
  files.forEach(file => {
    const inputPath = path.join(inputDir, file);
    const stat = fs.statSync(inputPath);
    
    if (stat.isDirectory()) {
      // Skip subdirectories for now
      return;
    }
    
    if (file.endsWith('.md') && file !== 'README.md') {
      const outputPath = path.join(outputDir, file.replace('.md', '.html'));
      convertMarkdownFile(inputPath, outputPath, category);
    }
  });
};

// Main conversion process
const main = () => {
  const docsSource = process.env.DOCS_SOURCE_DIR || './docs';
  const outputBase = process.env.OUTPUT_DIR || './public';
  const outputDocs = path.join(outputBase, 'docs');
  
  console.log('Converting markdown documentation to HTML...');
  console.log(`Source: ${docsSource}`);
  console.log(`Output: ${outputDocs}`);
  
  // Convert each category
  const categories = [
    { dir: 'specs', title: 'Technical Specifications' },
    { dir: 'implementation', title: 'Implementation Documentation' },
    { dir: 'user', title: 'User Guides' }
  ];
  
  categories.forEach(({ dir, title }) => {
    const inputDir = path.join(docsSource, dir);
    const outputDir = path.join(outputDocs, dir);
    convertDirectory(inputDir, outputDir, title);
  });
  
  console.log('Conversion complete!');
};

// Run if called directly
if (require.main === module) {
  main();
}

module.exports = { convertMarkdownFile, convertDirectory };

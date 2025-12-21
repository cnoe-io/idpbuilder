#!/usr/bin/env node

/**
 * Generate individual example pages with syntax highlighting
 */

const fs = require('fs');
const path = require('path');

// Example definitions
const examples = [
    {
        filename: 'platform-simple.html',
        title: 'Simple Platform',
        description: 'A basic Platform CR that references a GiteaProvider. This is the minimal configuration needed to get started with IDP Builder.',
        complexity: 'Beginner',
        components: 'Platform CR, GiteaProvider reference',
        yamlFile: 'platform-simple.yaml',
        yamlPath: '../../../examples/platform-simple.yaml',
        overview: 'This example demonstrates the simplest possible Platform configuration. It creates a Platform CR that references an existing GiteaProvider.',
        useCases: [
            'Getting started with IDP Builder',
            'Understanding the basic structure of Platform CRs',
            'Testing Platform controller functionality',
            'Building a foundation for more complex setups'
        ],
        prerequisites: [
            'A Kubernetes cluster (e.g., created with <code>kind create cluster</code>)',
            'IDP Builder CRDs installed in your cluster',
            'A GiteaProvider CR named <code>my-gitea</code> in the <code>gitea</code> namespace'
        ],
        nextSteps: [
            '<a href="/docs/examples/platform-complete.html">Complete Platform</a> example for a full configuration',
            '<a href="/docs/examples/giteaprovider-simple.html">GiteaProvider configuration</a>',
            '<a href="/docs/examples/v1alpha2/platform-with-gateway.html">Platform with Gateway</a> for multi-component setups'
        ],
        active: true,
        section: 'examples'
    },
    {
        filename: 'platform-complete.html',
        title: 'Complete Platform',
        description: 'A complete example with both Platform and GiteaProvider CRs in a single file. This shows the full configuration with all components together.',
        complexity: 'Intermediate',
        components: 'Platform CR, GiteaProvider CR (inline)',
        yamlFile: 'platform-complete.yaml',
        yamlPath: '../../../examples/platform-complete.yaml',
        overview: 'This example shows a complete configuration that includes both the GiteaProvider and Platform CRs in a single YAML file, demonstrating the full stack.',
        useCases: [
            'Deploying everything in one go',
            'Understanding the relationship between Platform and providers',
            'Creating a complete development environment',
            'Learning the full Platform specification'
        ],
        prerequisites: [
            'A Kubernetes cluster running',
            'IDP Builder CRDs installed',
            'kubectl configured to access your cluster'
        ],
        nextSteps: [
            '<a href="/docs/examples/v1alpha2/platform-with-gateway.html">V1Alpha2 examples</a> for the modular architecture',
            '<a href="/docs/specs/controller-architecture-spec.html">V2 Controller Architecture</a> documentation'
        ],
        active: false,
        section: 'examples'
    },
    {
        filename: 'giteaprovider-simple.html',
        title: 'Simple GiteaProvider',
        description: 'A basic GiteaProvider CR with auto-generated credentials and organizations. This is the easiest way to get a Git server running.',
        complexity: 'Beginner',
        components: 'GiteaProvider CR',
        yamlFile: 'giteaprovider-simple.yaml',
        yamlPath: '../../../examples/giteaprovider-simple.yaml',
        overview: 'This example demonstrates a standalone GiteaProvider configuration with auto-generated credentials and pre-configured organizations.',
        useCases: [
            'Quick Git provider setup',
            'Testing Gitea functionality',
            'Development environments',
            'Learning GiteaProvider configuration'
        ],
        prerequisites: [
            'A Kubernetes cluster',
            'IDP Builder CRDs installed',
            'Sufficient cluster resources (see <a href="/docs/user/minimum-requirements.html">minimum requirements</a>)'
        ],
        nextSteps: [
            '<a href="/docs/examples/platform-simple.html">Simple Platform</a> to reference this provider',
            '<a href="/docs/examples/platform-complete.html">Complete Platform</a> for integrated setup'
        ],
        active: false,
        section: 'examples'
    },
    {
        filename: 'platform-with-gateway.html',
        title: 'Platform with Gateway',
        description: 'A v1alpha2 Platform CR that includes both GiteaProvider and NginxGateway, demonstrating the modular provider-based architecture.',
        complexity: 'Intermediate',
        components: 'Platform CR, GiteaProvider reference, NginxGateway reference',
        yamlFile: 'platform-with-gateway.yaml',
        yamlPath: '../../../examples/v1alpha2/platform-with-gateway.yaml',
        overview: 'This v1alpha2 example demonstrates the modular architecture where platform components are managed through separate Custom Resources.',
        useCases: [
            'Multi-component platform setups',
            'Understanding v1alpha2 architecture',
            'Production-ready configurations',
            'Gateway integration with Platform'
        ],
        prerequisites: [
            'Kubernetes cluster',
            'IDP Builder CRDs installed',
            'Both <code>gitea-local</code> GiteaProvider and <code>nginx-gateway</code> NginxGateway CRs created'
        ],
        nextSteps: [
            '<a href="/docs/examples/v1alpha2/giteaprovider.html">GiteaProvider</a> configuration details',
            '<a href="/docs/examples/v1alpha2/nginxgateway.html">NginxGateway</a> configuration details',
            '<a href="/docs/specs/controller-architecture-spec.html">V2 Architecture</a> documentation'
        ],
        active: true,
        section: 'v1alpha2'
    },
    {
        filename: 'giteaprovider.html',
        title: 'V1Alpha2 GiteaProvider',
        description: 'A v1alpha2 GiteaProvider configuration for the modular architecture.',
        complexity: 'Beginner',
        components: 'GiteaProvider CR',
        yamlFile: 'giteaprovider.yaml',
        yamlPath: '../../../examples/v1alpha2/giteaprovider.yaml',
        overview: 'This v1alpha2 GiteaProvider example shows the provider configuration that can be referenced by a Platform CR.',
        useCases: [
            'Standalone Git provider',
            'Component of larger platform',
            'Development environments',
            'GitOps workflows'
        ],
        prerequisites: [
            'Kubernetes cluster',
            'IDP Builder v1alpha2 CRDs installed'
        ],
        nextSteps: [
            '<a href="/docs/examples/v1alpha2/platform-with-gateway.html">Platform with Gateway</a> to use this provider',
            '<a href="/docs/examples/v1alpha2/nginxgateway.html">NginxGateway</a> for ingress configuration'
        ],
        active: false,
        section: 'v1alpha2'
    },
    {
        filename: 'nginxgateway.html',
        title: 'V1Alpha2 NginxGateway',
        description: 'A v1alpha2 NginxGateway configuration for ingress controller management.',
        complexity: 'Beginner',
        components: 'NginxGateway CR',
        yamlFile: 'nginxgateway.yaml',
        yamlPath: '../../../examples/v1alpha2/nginxgateway.yaml',
        overview: 'This v1alpha2 NginxGateway example shows how to configure an Nginx Ingress Controller as part of the platform.',
        useCases: [
            'Ingress controller setup',
            'Gateway management',
            'Service exposure',
            'Platform networking'
        ],
        prerequisites: [
            'Kubernetes cluster',
            'IDP Builder v1alpha2 CRDs installed'
        ],
        nextSteps: [
            '<a href="/docs/examples/v1alpha2/platform-with-gateway.html">Platform with Gateway</a> to integrate this gateway',
            '<a href="/docs/examples/v1alpha2/giteaprovider.html">GiteaProvider</a> for Git provider setup'
        ],
        active: false,
        section: 'v1alpha2'
    }
];

const createPageTemplate = (example) => {
    // Read the YAML content
    const yamlFilePath = example.section === 'v1alpha2' ?
        path.join(__dirname, '..', 'examples/v1alpha2', example.yamlFile) :
        path.join(__dirname, '..', 'examples', example.yamlFile);
    const yamlContent = fs.readFileSync(yamlFilePath, 'utf8');
    
    const outputDir = example.section === 'v1alpha2' ? 
        path.join(__dirname, '..', 'site/docs/examples/v1alpha2') :
        path.join(__dirname, '..', 'site/docs/examples');
    
    const breadcrumb = example.section === 'v1alpha2' ?
        '<a href="/docs">Documentation</a> / <a href="/docs/examples.html">Examples</a> / <a href="/docs/examples/v1alpha2/README.html">V1Alpha2</a> / ' + example.title :
        '<a href="/docs">Documentation</a> / <a href="/docs/examples.html">Examples</a> / ' + example.title;
    
    const downloadPath = example.section === 'v1alpha2' ?
        `/docs/examples/v1alpha2/${example.yamlFile}` :
        `/docs/examples/${example.yamlFile}`;

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="${example.title} - IDP Builder Examples">
    <title>${example.title} | IDP Builder Examples</title>
    <link rel="stylesheet" href="/css/style.css">
    <!-- Prism.js for syntax highlighting -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.css">
    <style>
        .docs-container {
            display: grid;
            grid-template-columns: 250px 1fr;
            gap: 2rem;
            margin-top: 2rem;
        }
        .docs-sidebar {
            position: sticky;
            top: 80px;
            height: fit-content;
        }
        .docs-sidebar h3 {
            font-size: 0.9rem;
            text-transform: uppercase;
            color: var(--text-muted);
            margin-top: 1.5rem;
            margin-bottom: 0.5rem;
            padding-left: 0.5rem;
        }
        .docs-sidebar h3:first-child {
            margin-top: 0;
        }
        .docs-sidebar ul {
            list-style: none;
            padding: 0;
        }
        .docs-sidebar li {
            margin-bottom: 0.5rem;
        }
        .docs-sidebar a {
            text-decoration: none;
            color: var(--text-color);
            padding: 0.5rem;
            display: block;
            border-radius: 4px;
            transition: background-color 0.3s;
            font-size: 0.95rem;
        }
        .docs-sidebar a:hover,
        .docs-sidebar a.active {
            background-color: var(--bg-alt);
            color: var(--primary-color);
        }
        .docs-sidebar details {
            margin-bottom: 0.5rem;
        }
        .docs-sidebar summary {
            cursor: pointer;
            padding: 0.5rem;
            font-size: 0.9rem;
            text-transform: uppercase;
            color: var(--text-muted);
            user-select: none;
            border-radius: 4px;
            transition: background-color 0.3s;
        }
        .docs-sidebar summary:hover {
            background-color: var(--bg-alt);
        }
        .docs-sidebar details[open] > summary {
            background-color: var(--bg-alt);
            margin-bottom: 0.5rem;
        }
        .docs-sidebar details ul {
            margin-left: 0;
            padding-left: 1rem;
        }
        .docs-content {
            max-width: 900px;
            width: 100%;
        }
        .docs-content h1 {
            margin-bottom: 1rem;
        }
        .docs-content h2 {
            margin-top: 2rem;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid var(--border-color);
        }
        .docs-content h3 {
            margin-top: 1.5rem;
            margin-bottom: 0.75rem;
        }
        .docs-content p {
            margin-bottom: 1rem;
            line-height: 1.6;
        }
        .docs-content ul, .docs-content ol {
            margin-bottom: 1rem;
            padding-left: 2rem;
            line-height: 1.6;
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
        .example-meta {
            background-color: var(--bg-alt);
            padding: 1rem;
            border-radius: 5px;
            margin-bottom: 2rem;
        }
        .example-meta h3 {
            margin-top: 0;
            margin-bottom: 0.5rem;
        }
        .example-meta p {
            margin-bottom: 0.5rem;
        }
        .download-btn {
            display: inline-block;
            padding: 0.5rem 1rem;
            background-color: var(--primary-color);
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin-top: 0.5rem;
            transition: background-color 0.3s;
        }
        .download-btn:hover {
            background-color: var(--primary-hover);
        }
        pre[class*="language-"] {
            margin: 1.5rem 0;
            border-radius: 5px;
        }
        code[class*="language-"],
        pre[class*="language-"] {
            font-size: 0.9rem;
            line-height: 1.5;
        }
        @media (max-width: 768px) {
            .docs-container {
                grid-template-columns: 1fr;
            }
            .docs-sidebar {
                position: relative;
                top: 0;
                background-color: var(--bg-alt);
                padding: 1rem;
                border-radius: 5px;
                margin-bottom: 1rem;
            }
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
                <li><a href="/">Home</a></li>
                <li><a href="/docs">Docs</a></li>
                <li><a href="/docs/examples.html">Examples</a></li>
                <li><a href="https://github.com/cnoe-io/idpbuilder" target="_blank" rel="noopener">GitHub</a></li>
            </ul>
        </nav>
    </header>

    <main class="container">
        <div class="docs-container">
            <aside class="docs-sidebar">
                <h3>Documentation</h3>
                <ul>
                    <li><a href="/docs">Overview</a></li>
                </ul>
                
                <details ${example.section === 'examples' ? 'open' : ''}>
                    <summary>Examples</summary>
                    <ul>
                        <li><a href="/docs/examples.html">Examples Overview</a></li>
                        <li><a href="/docs/examples/platform-simple.html" ${example.filename === 'platform-simple.html' ? 'class="active"' : ''}>Simple Platform</a></li>
                        <li><a href="/docs/examples/platform-complete.html" ${example.filename === 'platform-complete.html' ? 'class="active"' : ''}>Complete Platform</a></li>
                        <li><a href="/docs/examples/giteaprovider-simple.html" ${example.filename === 'giteaprovider-simple.html' ? 'class="active"' : ''}>Simple GiteaProvider</a></li>
                    </ul>
                </details>
                
                <details ${example.section === 'v1alpha2' ? 'open' : ''}>
                    <summary>V1Alpha2 Examples</summary>
                    <ul>
                        <li><a href="/docs/examples/v1alpha2/platform-with-gateway.html" ${example.filename === 'platform-with-gateway.html' ? 'class="active"' : ''}>Platform with Gateway</a></li>
                        <li><a href="/docs/examples/v1alpha2/giteaprovider.html" ${example.filename === 'giteaprovider.html' ? 'class="active"' : ''}>GiteaProvider</a></li>
                        <li><a href="/docs/examples/v1alpha2/nginxgateway.html" ${example.filename === 'nginxgateway.html' ? 'class="active"' : ''}>NginxGateway</a></li>
                    </ul>
                </details>
                
                <details>
                    <summary>Technical Specs</summary>
                    <ul>
                        <li><a href="/docs/specs/controller-architecture-spec.html">V2 Controller Architecture</a></li>
                        <li><a href="/docs/specs/hyperscaler-provider-spec.html">Hyperscaler Provider</a></li>
                        <li><a href="/docs/specs/pluggable-packages.html">Pluggable Packages</a></li>
                    </ul>
                </details>
                
                <details>
                    <summary>User Guides</summary>
                    <ul>
                        <li><a href="/docs/user/minimum-requirements.html">Minimum Requirements</a></li>
                        <li><a href="/docs/user/private-registries.html">Private Registries</a></li>
                    </ul>
                </details>
            </aside>

            <article class="docs-content">
                <div class="breadcrumb">
                    ${breadcrumb}
                </div>
                
                <h1>${example.title}</h1>
                
                <div class="example-meta">
                    <h3>📋 About This Example</h3>
                    <p><strong>Use case:</strong> ${example.description}</p>
                    <p><strong>Complexity:</strong> ${example.complexity}</p>
                    <p><strong>Components:</strong> ${example.components}</p>
                    <a href="${downloadPath}" class="download-btn" download>⬇ Download YAML</a>
                </div>

                <h2>Overview</h2>
                <p>${example.overview}</p>
                <p>This configuration is perfect for:</p>
                <ul>
                    ${example.useCases.map(uc => `<li>${uc}</li>`).join('\n                    ')}
                </ul>

                <h2>YAML Configuration</h2>
                <pre class="line-numbers"><code class="language-yaml">${yamlContent.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</code></pre>

                <h2>Prerequisites</h2>
                <p>Before applying this configuration, ensure you have:</p>
                <ul>
                    ${example.prerequisites.map(p => `<li>${p}</li>`).join('\n                    ')}
                </ul>

                <h2>Usage</h2>
                <h3>1. Apply the manifest</h3>
                <pre><code class="language-bash">kubectl apply -f ${example.yamlFile}</code></pre>

                <h3>2. Check the status</h3>
                <pre><code class="language-bash"># View resources
kubectl get ${example.yamlFile.includes('platform') ? 'platform' : example.yamlFile.includes('gitea') ? 'giteaprovider' : 'nginxgateway'}

# Get detailed information
kubectl describe ${example.yamlFile.includes('platform') ? 'platform' : example.yamlFile.includes('gitea') ? 'giteaprovider' : 'nginxgateway'}</code></pre>

                <h2>Next Steps</h2>
                <p>After deploying this example, you can:</p>
                <ul>
                    ${example.nextSteps.map(ns => `<li>Explore the ${ns}</li>`).join('\n                    ')}
                </ul>
            </article>
        </div>
    </main>

    <footer>
        <div class="container">
            <p>&copy; 2024 CNOE IDP Builder. Licensed under <a href="https://github.com/cnoe-io/idpbuilder/blob/main/LICENSE" target="_blank" rel="noopener">Apache License 2.0</a></p>
        </div>
    </footer>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-core.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/autoloader/prism-autoloader.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.js"></script>
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
</html>
`;
};

// Generate all example pages
examples.forEach(example => {
    const html = createPageTemplate(example);
    const outputDir = example.section === 'v1alpha2' ? 
        path.join(__dirname, '..', 'site/docs/examples/v1alpha2') :
        path.join(__dirname, '..', 'site/docs/examples');
    
    const outputPath = path.join(outputDir, example.filename);
    fs.writeFileSync(outputPath, html);
    console.log(`Generated: ${outputPath}`);
});

console.log('All example pages generated successfully!');

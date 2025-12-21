# Scripts

This directory contains utility scripts for the idpbuilder project, including test analysis and documentation validation tools.

## analyze_test_times.py

A Python script that analyzes test execution times and generates visualizations showing which tests take the longest to run and why.

### Usage

First, run the tests with JSON output:

```bash
go test --tags=integration -v -timeout 30m ./... -json 2>&1 | tee test-output.json
```

Then analyze the results:

```bash
python3 scripts/analyze_test_times.py test-output.json docs/implementation/test-timing-analysis.md
```

This will:
1. Display a text-based diagram in the console
2. Save a detailed markdown report to the specified file

### Features

- **Test Categorization**: Automatically categorizes tests into:
  - Integration tests (controller reconciliation logic)
  - I/O tests (file system, git operations)
  - Build/Manifest tests (Kubernetes manifest generation)
  - Config/Validation tests (configuration parsing)
  - Unit tests (fast, isolated tests)

- **Visual Reports**: Generates ASCII bar charts and markdown tables showing:
  - Slowest individual tests
  - Time distribution by category
  - Time distribution by package

- **Analysis**: Provides insights into why tests take long and recommendations for improvement

### Output

The script generates two outputs:

1. **Console Output**: A text-based diagram showing test timing statistics
2. **Markdown Report**: A detailed report with tables, charts, and analysis (when output file is specified)

See [docs/implementation/test-timing-analysis.md](../docs/implementation/test-timing-analysis.md) for an example of the generated report.

## validate-docs-sync.py

A Python script that validates documentation files in the `docs/` directories are properly linked in the site navigation (`site/docs/index.html`).

### Usage

```bash
# Run validation directly
python3 scripts/validate-docs-sync.py

# Or use the Makefile target
make validate-docs
```

### Features

- **Automatic Discovery**: Scans `docs/specs`, `docs/implementation`, and `docs/user` directories for markdown files
- **Navigation Parsing**: Extracts links from the site navigation HTML
- **Validation**: Compares actual files against navigation links and reports discrepancies
- **Exit Codes**: Returns 0 if all files are linked, 1 if missing links are found (suitable for CI/CD)

### Output

The script provides a clear report showing:
- ✅ Categories where all files are linked
- ❌ Categories with missing links
- Specific filenames that need to be added to navigation

### Integration

This validation is designed to be integrated into CI/CD workflows to ensure documentation stays in sync with navigation. Add it to the PR workflow to catch missing links before they're merged:

```yaml
- name: Validate docs sync
  run: make validate-docs
```

### Why This Matters

Keeping documentation files synced with site navigation ensures:
1. All documentation is discoverable by users
2. No orphaned documentation files exist
3. The documentation site provides a complete view of available docs
4. Documentation updates are immediately visible on the site

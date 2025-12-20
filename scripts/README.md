# Test Analysis Scripts

This directory contains scripts for analyzing test execution times in the idpbuilder project.

## analyze_test_times.py

A Python script that analyzes test execution times and generates visualizations showing which tests take the longest to run and why.

### Usage

First, run the tests with JSON output:

```bash
go test --tags=integration -v -timeout 30m ./... -json 2>&1 | tee test-output.json
```

Then analyze the results:

```bash
python3 scripts/analyze_test_times.py test-output.json docs/test-timing-analysis.md
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

See [docs/test-timing-analysis.md](../docs/test-timing-analysis.md) for an example of the generated report.

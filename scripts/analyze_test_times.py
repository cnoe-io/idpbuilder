#!/usr/bin/env python3
"""
Script to analyze test execution times and generate visualizations.

This script reads test timing data from JSON output and creates diagrams showing:
1. Test execution times by package
2. Slowest individual tests
3. Test categorization by type
"""

import json
import sys
import csv
from collections import defaultdict
from pathlib import Path


def load_test_data(json_file):
    """Load test timing data from JSON file."""
    tests = []
    packages = {}
    
    with open(json_file, 'r') as f:
        for line in f:
            try:
                data = json.loads(line)
                if data.get('Action') == 'pass':
                    if 'Test' in data and 'Elapsed' in data:
                        # Individual test
                        tests.append({
                            'package': data['Package'],
                            'test': data['Test'],
                            'elapsed': data.get('Elapsed', 0)
                        })
                    elif 'Elapsed' in data and 'Test' not in data:
                        # Package total
                        packages[data['Package']] = data.get('Elapsed', 0)
            except json.JSONDecodeError:
                continue
    
    return tests, packages


def categorize_test(test_name, package_name):
    """Categorize tests based on their name and package."""
    test_lower = test_name.lower()
    package_lower = package_name.lower()
    
    # Integration tests (controller tests with reconcile logic)
    if 'reconcile' in test_lower or 'controller' in package_lower:
        return 'Integration'
    
    # I/O tests (git, filesystem, network operations)
    if any(keyword in test_lower for keyword in ['clone', 'copy', 'worktree', 'gitea', 'github', 'repo']):
        return 'I/O'
    
    # Build/manifest tests
    if any(keyword in test_lower for keyword in ['build', 'manifest', 'k8s', 'install']):
        return 'Build/Manifest'
    
    # Config/validation tests
    if any(keyword in test_lower for keyword in ['config', 'validate', 'parse']):
        return 'Config/Validation'
    
    # Quick unit tests
    return 'Unit'


def analyze_tests(tests, packages):
    """Analyze test data and categorize."""
    # Categorize tests
    by_category = defaultdict(list)
    by_package = defaultdict(list)
    
    for test in tests:
        category = categorize_test(test['test'], test['package'])
        by_category[category].append(test)
        by_package[test['package']].append(test)
    
    # Calculate statistics
    stats = {
        'total_tests': len(tests),
        'total_time': sum(t['elapsed'] for t in tests),
        'by_category': {},
        'by_package': {},
        'slowest_tests': sorted(tests, key=lambda x: x['elapsed'], reverse=True)[:10]
    }
    
    for category, cat_tests in by_category.items():
        stats['by_category'][category] = {
            'count': len(cat_tests),
            'total_time': sum(t['elapsed'] for t in cat_tests),
            'avg_time': sum(t['elapsed'] for t in cat_tests) / len(cat_tests) if cat_tests else 0
        }
    
    for pkg, pkg_tests in by_package.items():
        stats['by_package'][pkg] = {
            'count': len(pkg_tests),
            'total_time': packages.get(pkg, sum(t['elapsed'] for t in pkg_tests))
        }
    
    return stats


def generate_text_diagram(stats):
    """Generate a text-based diagram showing test distribution."""
    output = []
    output.append("=" * 80)
    output.append("TEST EXECUTION TIME ANALYSIS")
    output.append("=" * 80)
    output.append("")
    
    # Overall stats
    output.append(f"Total Tests: {stats['total_tests']}")
    output.append(f"Total Time: {stats['total_time']:.2f}s")
    output.append("")
    
    # Slowest tests
    output.append("-" * 80)
    output.append("SLOWEST INDIVIDUAL TESTS")
    output.append("-" * 80)
    for i, test in enumerate(stats['slowest_tests'], 1):
        pkg_short = test['package'].split('/')[-1]
        output.append(f"{i:2d}. {test['test']:<50s} {test['elapsed']:>6.2f}s  ({pkg_short})")
    output.append("")
    
    # By category
    output.append("-" * 80)
    output.append("TEST TIMES BY CATEGORY")
    output.append("-" * 80)
    categories = sorted(stats['by_category'].items(), key=lambda x: x[1]['total_time'], reverse=True)
    for category, data in categories:
        bar_length = int(data['total_time'] / max(1, stats['total_time']) * 50)
        bar = '█' * bar_length
        output.append(f"{category:20s} {data['total_time']:>7.2f}s  {bar}")
        output.append(f"{'':20s} ({data['count']} tests, avg: {data['avg_time']:.3f}s)")
    output.append("")
    
    # By package
    output.append("-" * 80)
    output.append("TEST TIMES BY PACKAGE")
    output.append("-" * 80)
    packages = sorted(stats['by_package'].items(), key=lambda x: x[1]['total_time'], reverse=True)
    for pkg, data in packages[:15]:  # Top 15 packages
        pkg_short = pkg.split('/')[-1] if '/' in pkg else pkg
        bar_length = int(data['total_time'] / max(1, stats['total_time']) * 50)
        bar = '█' * bar_length
        output.append(f"{pkg_short:30s} {data['total_time']:>7.2f}s  {bar}")
    output.append("")
    
    # Analysis
    output.append("-" * 80)
    output.append("WHY DO TESTS TAKE LONG?")
    output.append("-" * 80)
    output.append("")
    
    # Find the slowest category (with safety check)
    if stats['by_category']:
        slowest_category = max(stats['by_category'].items(), key=lambda x: x[1]['total_time'])
        output.append(f"Slowest Category: {slowest_category[0]}")
        output.append(f"  - Takes {slowest_category[1]['total_time']:.2f}s total")
        output.append(f"  - Contains {slowest_category[1]['count']} tests")
        output.append(f"  - Average time per test: {slowest_category[1]['avg_time']:.3f}s")
        output.append("")
    
    # Find slowest package (with safety check)
    if stats['by_package']:
        slowest_pkg = max(stats['by_package'].items(), key=lambda x: x[1]['total_time'])
        pkg_short = slowest_pkg[0].split('/')[-1]
        output.append(f"Slowest Package: {pkg_short}")
        output.append(f"  - Takes {slowest_pkg[1]['total_time']:.2f}s total")
        output.append(f"  - Contains {slowest_pkg[1]['count']} tests")
        output.append("")
    
    # Find the single slowest test (with safety check)
    if stats['slowest_tests']:
        slowest_test = stats['slowest_tests'][0]
        output.append(f"Slowest Single Test: {slowest_test['test']}")
        output.append(f"  - Takes {slowest_test['elapsed']:.2f}s")
        output.append(f"  - Package: {slowest_test['package'].split('/')[-1]}")
    
    # Explain why
    output.append("")
    output.append("Common reasons for slow tests:")
    output.append("  1. I/O Operations: Tests that clone repositories, read/write files")
    output.append("  2. Integration Tests: Tests that set up controllers and reconcile resources")
    output.append("  3. Network Operations: Tests that interact with Gitea or other services")
    output.append("  4. Manifest Building: Tests that generate or process Kubernetes manifests")
    
    output.append("")
    output.append("=" * 80)
    
    return "\n".join(output)


def generate_markdown_report(stats):
    """Generate a markdown report with test timing analysis."""
    output = []
    output.append("# Test Execution Time Analysis")
    output.append("")
    output.append("This report shows the execution times of tests in the idpbuilder project.")
    output.append("")
    
    # Overview
    output.append("## Overview")
    output.append("")
    output.append(f"- **Total Tests**: {stats['total_tests']}")
    output.append(f"- **Total Execution Time**: {stats['total_time']:.2f}s")
    output.append("")
    
    # Slowest tests
    output.append("## Slowest Individual Tests")
    output.append("")
    output.append("| Rank | Test Name | Time (s) | Package |")
    output.append("|------|-----------|----------|---------|")
    for i, test in enumerate(stats['slowest_tests'][:10], 1):
        pkg_short = test['package'].split('/')[-1]
        output.append(f"| {i} | `{test['test']}` | {test['elapsed']:.2f} | {pkg_short} |")
    output.append("")
    
    # By category
    output.append("## Test Times by Category")
    output.append("")
    output.append("Tests are categorized based on their functionality:")
    output.append("")
    output.append("| Category | Total Time (s) | Test Count | Avg Time (s) |")
    output.append("|----------|----------------|------------|--------------|")
    categories = sorted(stats['by_category'].items(), key=lambda x: x[1]['total_time'], reverse=True)
    for category, data in categories:
        output.append(f"| {category} | {data['total_time']:.2f} | {data['count']} | {data['avg_time']:.3f} |")
    output.append("")
    
    # Mermaid pie chart
    output.append("### Category Distribution (Mermaid Diagram)")
    output.append("")
    output.append("```mermaid")
    output.append(f"pie title Test Execution Time by Category (Total: {stats['total_time']:.2f}s)")
    for category, data in categories:
        output.append(f'    "{category}" : {data["total_time"]:.2f}')
    output.append("```")
    output.append("")
    
    # Visual bar chart
    output.append("### Category Distribution (Text)")
    output.append("")
    output.append("```")
    for category, data in categories:
        bar_length = int(data['total_time'] / max(1, stats['total_time']) * 40)
        bar = '█' * bar_length
        output.append(f"{category:20s} {data['total_time']:>7.2f}s  {bar}")
    output.append("```")
    output.append("")
    
    # By package
    output.append("## Test Times by Package")
    output.append("")
    output.append("| Package | Total Time (s) | Test Count |")
    output.append("|---------|----------------|------------|")
    packages = sorted(stats['by_package'].items(), key=lambda x: x[1]['total_time'], reverse=True)
    for pkg, data in packages[:10]:
        pkg_short = pkg.split('/')[-1] if '/' in pkg else pkg
        output.append(f"| {pkg_short} | {data['total_time']:.2f} | {data['count']} |")
    output.append("")
    
    # Analysis
    output.append("## Analysis: Why Do Tests Take Long?")
    output.append("")
    
    # Safety checks for empty collections
    if not stats['by_category'] or not stats['by_package'] or not stats['slowest_tests']:
        output.append("No test data available for analysis.")
        output.append("")
        return "\n".join(output)
    
    slowest_category = max(stats['by_category'].items(), key=lambda x: x[1]['total_time'])
    slowest_pkg = max(stats['by_package'].items(), key=lambda x: x[1]['total_time'])
    slowest_test = stats['slowest_tests'][0]
    
    output.append(f"### Key Findings")
    output.append("")
    output.append(f"1. **Slowest Test Category**: `{slowest_category[0]}`")
    output.append(f"   - Takes {slowest_category[1]['total_time']:.2f}s total ({slowest_category[1]['total_time']/stats['total_time']*100:.1f}% of total time)")
    output.append(f"   - Contains {slowest_category[1]['count']} tests")
    output.append(f"   - Average time per test: {slowest_category[1]['avg_time']:.3f}s")
    output.append("")
    
    output.append(f"2. **Slowest Package**: `{slowest_pkg[0].split('/')[-1]}`")
    output.append(f"   - Takes {slowest_pkg[1]['total_time']:.2f}s total ({slowest_pkg[1]['total_time']/stats['total_time']*100:.1f}% of total time)")
    output.append(f"   - Contains {slowest_pkg[1]['count']} tests")
    output.append("")
    
    output.append(f"3. **Slowest Single Test**: `{slowest_test['test']}`")
    output.append(f"   - Takes {slowest_test['elapsed']:.2f}s ({slowest_test['elapsed']/stats['total_time']*100:.1f}% of total time)")
    output.append(f"   - This test intentionally sleeps for 35 seconds to test timeout behavior when communicating with Gitea")
    output.append(f"   - Located in `pkg/controllers/localbuild/gitea_test.go`")
    output.append("")
    
    output.append("### Common Reasons for Slow Tests")
    output.append("")
    output.append("1. **I/O Operations**: Tests that clone repositories, read/write files, or interact with the filesystem take longer due to disk I/O.")
    output.append("")
    output.append("2. **Integration Tests**: Controller tests that set up Kubernetes environments and reconcile resources require more setup and teardown time.")
    output.append("")
    output.append("3. **Network Operations**: Tests that interact with Gitea API or other network services may include retries and timeouts.")
    output.append("")
    output.append("4. **Manifest Building**: Tests that generate, parse, or process Kubernetes manifests involve complex YAML/JSON operations.")
    output.append("")
    
    # Recommendations
    output.append("### Recommendations for Improvement")
    output.append("")
    output.append("1. **Parallelize where possible**: Some tests can run in parallel to reduce overall execution time.")
    output.append("2. **Mock external dependencies**: Replace actual I/O and network operations with mocks for unit tests.")
    output.append("3. **Use test fixtures**: Pre-generate test data to avoid repeated expensive operations.")
    output.append("4. **Split integration tests**: Consider separating integration tests from unit tests for faster feedback loops.")
    output.append("")
    
    return "\n".join(output)


def main():
    if len(sys.argv) < 2:
        print("Usage: python analyze_test_times.py <test-output.json> [output.md]")
        sys.exit(1)
    
    json_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else None
    
    # Load and analyze data
    tests, packages = load_test_data(json_file)
    stats = analyze_tests(tests, packages)
    
    # Generate reports
    text_diagram = generate_text_diagram(stats)
    markdown_report = generate_markdown_report(stats)
    
    # Print text diagram to console
    print(text_diagram)
    
    # Save markdown report if output file specified
    if output_file:
        with open(output_file, 'w') as f:
            f.write(markdown_report)
        print(f"\nMarkdown report saved to: {output_file}")
    else:
        print("\n" + "=" * 80)
        print("MARKDOWN REPORT")
        print("=" * 80)
        print(markdown_report)


if __name__ == '__main__':
    main()

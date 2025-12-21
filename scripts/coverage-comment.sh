#!/usr/bin/env bash
set -euo pipefail

# This script generates a coverage report from cover.out and posts it as a PR comment
# Requires: gh (GitHub CLI)
# Environment variables: GH_TOKEN, PR_NUMBER

COVERAGE_FILE="${1:-cover.out}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "Error: Coverage file '$COVERAGE_FILE' not found"
    exit 1
fi

# Check if we're in a PR context
if [ -z "${PR_NUMBER:-}" ]; then
    echo "Not in a PR context (PR_NUMBER not set). Skipping comment."
    exit 0
fi

# Get repository from git remote (owner/repo format)
REPO=$(git remote get-url origin 2>/dev/null | sed -E 's#https://github.com/##; s#git@github.com:##; s#\.git$##')
if [ -z "$REPO" ]; then
    REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo "none")
    echo "Error: Could not determine repository from git remote. Remote URL: $REMOTE_URL"
    exit 1
fi

# Generate coverage report
echo "Generating coverage report from $COVERAGE_FILE..."

# Parse coverage file and calculate statistics
# Format: path/to/file.go:start.col,end.col num_statements count
# We calculate: (lines with count > 0) / (total lines) * 100

COVERAGE_STATS=$(awk '
    NR == 1 { next }  # Skip mode line
    {
        # Extract package path from filename (everything except the filename)
        split($1, parts, ":")
        filepath = parts[1]
        
        # Get package from filepath
        n = split(filepath, pathparts, "/")
        if (n > 1) {
            pkg = ""
            for (i = 1; i < n; i++) {
                if (i > 1) pkg = pkg "/"
                pkg = pkg pathparts[i]
            }
        } else {
            pkg = "."
        }
        
        # Track statistics
        statements = $2
        covered = $3
        
        total_statements += statements
        if (covered > 0) {
            covered_statements += statements
        }
        
        pkg_total[pkg] += statements
        if (covered > 0) {
            pkg_covered[pkg] += statements
        }
    }
    END {
        # Calculate total coverage
        if (total_statements > 0) {
            total_pct = (covered_statements / total_statements) * 100
        } else {
            total_pct = 0
        }
        
        print "TOTAL:" total_pct
        
        # Print per-package coverage
        for (pkg in pkg_total) {
            if (pkg_total[pkg] > 0) {
                pkg_pct = (pkg_covered[pkg] / pkg_total[pkg]) * 100
            } else {
                pkg_pct = 0
            }
            print pkg ":" pkg_pct
        }
    }
' "$COVERAGE_FILE")

# Extract total coverage
TOTAL_COVERAGE=$(echo "$COVERAGE_STATS" | grep "^TOTAL:" | cut -d: -f2)
TOTAL_COVERAGE_FORMATTED=$(printf "%.1f%%" "$TOTAL_COVERAGE")

# Extract per-package coverage and format as table
PACKAGE_COVERAGE=$(echo "$COVERAGE_STATS" | grep -v "^TOTAL:" | sort | awk -F: '{
    printf "| `%s` | %.1f%% |\n", $1, $2
}')

# Create the comment body
COMMENT_BODY=$(cat <<EOF
## 📊 Test Coverage Report

**Total Coverage:** \`${TOTAL_COVERAGE_FORMATTED}\`

### Coverage by Package

| Package | Coverage |
|---------|----------|
${PACKAGE_COVERAGE}

---
*Coverage report generated from \`${COVERAGE_FILE}\`*
EOF
)

# Post comment to PR
echo "Posting coverage comment to PR #${PR_NUMBER} in ${REPO}..."
echo "$COMMENT_BODY" | gh pr comment "$PR_NUMBER" --repo "$REPO" --body-file -

echo "✅ Coverage comment posted successfully!"

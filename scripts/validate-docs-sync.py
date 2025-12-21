#!/usr/bin/env python3
"""
Validate that all documentation files in docs/ directories are linked in site navigation.

This script ensures that the documentation in docs/specs, docs/implementation, and docs/user
directories are properly linked in the site/docs/index.html navigation sidebar.

Exit codes:
  0: All documentation files are properly linked
  1: Missing links found or other errors
"""

import os
import re
import sys
from pathlib import Path
from typing import Dict, Set, List, Tuple


def get_markdown_files(base_dir: str, category: str) -> Set[str]:
    """
    Get all markdown files in a docs category directory.
    
    Args:
        base_dir: Base directory of the repository
        category: Documentation category (specs, implementation, user)
    
    Returns:
        Set of markdown filenames (excluding README.md)
    """
    docs_dir = os.path.join(base_dir, 'docs', category)
    if not os.path.exists(docs_dir):
        return set()
    
    md_files = set()
    for filename in os.listdir(docs_dir):
        if filename.endswith('.md') and filename != 'README.md':
            md_files.add(filename)
    
    return md_files


def extract_nav_links(html_file: str) -> Dict[str, Set[str]]:
    """
    Extract documentation links from the site navigation HTML.
    
    Expects links in the format: <a href="/docs/{category}/{file}.html">Title</a>
    where category is one of: specs, implementation, user
    
    The regex pattern handles standard HTML attributes but assumes well-formed HTML.
    It will not correctly handle escaped quotes or complex nested structures.
    
    Args:
        html_file: Path to site/docs/index.html
    
    Returns:
        Dictionary mapping category to set of linked markdown filenames
    """
    if not os.path.exists(html_file):
        print(f"Error: Navigation file not found: {html_file}", file=sys.stderr)
        sys.exit(1)
    
    with open(html_file, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Pattern to match links like: <a href="/docs/implementation/file.html">Title</a>
    # Captures: (category, filename, link_text)
    link_pattern = r'<a href="/docs/(specs|implementation|user)/([^"]+)">([^<]+)</a>'
    links_found = re.findall(link_pattern, content)
    
    # Store linked files by directory
    linked_files = {'specs': set(), 'implementation': set(), 'user': set()}
    
    for category, filename, text in links_found:
        # Convert .html to .md (assumes all links end with .html)
        if filename.endswith('.html'):
            md_file = filename.replace('.html', '.md')
            linked_files[category].add(md_file)
    
    return linked_files


def validate_docs_sync(base_dir: str) -> Tuple[bool, List[str]]:
    """
    Validate that all documentation files are linked in navigation.
    
    Args:
        base_dir: Base directory of the repository
    
    Returns:
        Tuple of (success: bool, messages: List[str])
    """
    html_file = os.path.join(base_dir, 'site', 'docs', 'index.html')
    categories = ['specs', 'implementation', 'user']
    
    # Get linked files from navigation
    linked_files = extract_nav_links(html_file)
    
    # Check each category for missing links
    all_good = True
    messages = []
    
    for category in categories:
        actual_files = get_markdown_files(base_dir, category)
        linked = linked_files[category]
        
        missing = actual_files - linked
        
        if missing:
            all_good = False
            messages.append(f"\n❌ {category.upper()}: Found {len(missing)} file(s) not linked in navigation:")
            for filename in sorted(missing):
                messages.append(f"   - {filename}")
        else:
            messages.append(f"✅ {category.upper()}: All {len(actual_files)} file(s) are linked")
    
    return all_good, messages


def main():
    """Main entry point."""
    # Determine base directory (repository root)
    script_dir = os.path.dirname(os.path.abspath(__file__))
    base_dir = os.path.dirname(script_dir)
    
    print("=" * 70)
    print("Documentation Sync Validation")
    print("=" * 70)
    print(f"Repository: {base_dir}")
    print(f"Checking: docs/specs, docs/implementation, docs/user")
    print(f"Against: site/docs/index.html")
    print("=" * 70)
    
    success, messages = validate_docs_sync(base_dir)
    
    # Print all messages
    for msg in messages:
        print(msg)
    
    print("=" * 70)
    
    if success:
        print("✅ SUCCESS: All documentation files are linked in navigation!")
        return 0
    else:
        print("❌ FAILURE: Some documentation files are missing from navigation!")
        print("\nTo fix this:")
        print("1. Edit site/docs/index.html")
        print("2. Add missing files to the appropriate <details> section")
        print("3. Use the pattern: <a href=\"/docs/CATEGORY/filename.html\">Title</a>")
        return 1


if __name__ == '__main__':
    sys.exit(main())

#!/usr/bin/env python3
"""
Parse ctags file with extended fields and generate vim-vista-like tree structure.
Usage: ./ctags-tree.py <tags_file> <target_file>
Example: ./ctags-tree.py tags ai-docs/planning/backlog/route-domain-simplification-refactoring.md
"""

import sys
import re
from collections import defaultdict
from typing import List, Dict, Tuple, Optional


class TagEntry:
    """Represents a single ctags entry."""

    def __init__(self, name: str, file: str, pattern: str, kind: str,
                 line: Optional[int], scope: Optional[str]):
        self.name = name
        self.file = file
        self.pattern = pattern
        self.kind = kind
        self.line = line
        self.scope = scope  # Full scope with "" separators
        self.level = self._determine_level(kind)

    def _determine_level(self, kind: str) -> int:
        """Map kind to heading level (H1-H6)."""
        kind_map = {
            'chapter': 1,       # H1: #
            'section': 2,       # H2: ##
            'subsection': 3,    # H3: ###
            'subsubsection': 4, # H4: ####
        }
        return kind_map.get(kind, 0)

    def __repr__(self):
        return f"TagEntry({self.name}, line={self.line}, level=H{self.level})"


def parse_tags_file(tags_path: str, target_file: str) -> List[TagEntry]:
    """Parse ctags file and extract entries for target file."""
    entries = []

    with open(tags_path, 'r', encoding='utf-8') as f:
        for line in f:
            # Skip meta lines starting with !_TAG
            if line.startswith('!_TAG'):
                continue

            # Parse tab-separated fields
            parts = line.rstrip('\n').split('\t')
            if len(parts) < 4:
                continue

            name = parts[0]
            file = parts[1]
            pattern = parts[2]

            # Only process entries for target file
            if file != target_file:
                continue

            # Parse the rest of the fields
            kind = None
            line_num = None
            scope = None

            for part in parts[3:]:
                if part.startswith(';"'):
                    # Skip extension marker
                    continue
                elif part in ['chapter', 'section', 'subsection', 'subsubsection']:
                    kind = part
                elif part.startswith('line:'):
                    line_num = int(part.split(':')[1])
                elif part.startswith('chapter:') or part.startswith('section:') or \
                     part.startswith('subsection:'):
                    # Extract scope (parent hierarchy)
                    scope = part.split(':', 1)[1] if ':' in part else None

            if kind:
                entry = TagEntry(name, file, pattern, kind, line_num, scope)
                entries.append(entry)

    # Sort by line number
    entries.sort(key=lambda e: e.line if e.line else 0)
    return entries


def build_tree_structure(entries: List[TagEntry]) -> str:
    """Build vim-vista-like tree structure from tag entries."""
    if not entries:
        return ""

    lines = []
    stack = []  # Track parent entries at each level

    for entry in entries:
        level = entry.level

        # Pop stack to current level
        while stack and stack[-1][0] >= level:
            stack.pop()

        # Calculate indentation
        indent = '  ' * len(stack)

        # Determine tree character
        if len(stack) == 0:
            tree_char = '└'
        else:
            tree_char = '│'

        # Format line
        line_info = f"H{level}:{entry.line}" if entry.line else f"H{level}"
        formatted = f"{indent}{tree_char} {entry.name} {line_info}"
        lines.append(formatted)

        # Push to stack for children
        stack.append((level, entry))

    # Add filename as root
    if entries:
        filename = entries[0].file.split('/')[-1]
        result = f"{filename}\n\n"
        result += '\n'.join(lines)
        return result

    return '\n'.join(lines)


def main():
    if len(sys.argv) < 3:
        print("Usage: ctags-tree.py <tags_file> <target_file>")
        print("Example: ctags-tree.py tags ai-docs/planning/backlog/route-domain-simplification-refactoring.md")
        sys.exit(1)

    tags_file = sys.argv[1]
    target_file = sys.argv[2]

    # Parse tags file
    entries = parse_tags_file(tags_file, target_file)

    if not entries:
        print(f"No entries found for {target_file}")
        sys.exit(1)

    # Build and print tree
    tree = build_tree_structure(entries)
    print(tree)


if __name__ == '__main__':
    main()

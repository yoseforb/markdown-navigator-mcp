#!/usr/bin/env python3
"""
Extract a specific section from a markdown file using ctags.
Usage: ./ctags-section.py <tags_file> <target_file> <section_name>
Example: ./ctags-section.py tags ai-docs/planning/backlog/route-domain-simplification-refactoring.md "Task 4"
"""

import sys
import re


def find_section_bounds(tags_file: str, target_file: str, section_query: str):
    """Find line number bounds for a section."""
    entries = []

    with open(tags_file, 'r', encoding='utf-8') as f:
        for line in f:
            if line.startswith('!_TAG'):
                continue

            parts = line.rstrip('\n').split('\t')
            if len(parts) < 4:
                continue

            name = parts[0]
            file = parts[1]

            if file != target_file:
                continue

            # Parse line number and kind
            line_num = None
            kind = None

            for part in parts[3:]:
                if part in ['chapter', 'section', 'subsection', 'subsubsection']:
                    kind = part
                elif part.startswith('line:'):
                    line_num = int(part.split(':')[1])

            if line_num and kind:
                entries.append((name, line_num, kind))

    # Sort by line number
    entries.sort(key=lambda e: e[1])

    # Find matching section
    target_idx = None
    target_kind = None
    for idx, (name, line_num, kind) in enumerate(entries):
        if section_query.lower() in name.lower():
            target_idx = idx
            target_kind = kind
            break

    if target_idx is None:
        return None, None, None

    start_line = entries[target_idx][1]
    section_name = entries[target_idx][0]

    # Find end line (next section at same or higher level)
    kind_levels = {
        'chapter': 1,
        'section': 2,
        'subsection': 3,
        'subsubsection': 4
    }
    target_level = kind_levels.get(target_kind, 2)

    end_line = None
    for idx in range(target_idx + 1, len(entries)):
        _, line_num, kind = entries[idx]
        if kind_levels.get(kind, 5) <= target_level:
            end_line = line_num - 1
            break

    return start_line, end_line, section_name


def main():
    if len(sys.argv) < 4:
        print("Usage: ctags-section.py <tags_file> <target_file> <section_query>")
        print('Example: ctags-section.py tags route-domain-simplification-refactoring.md "Task 4"')
        sys.exit(1)

    tags_file = sys.argv[1]
    target_file = sys.argv[2]
    section_query = sys.argv[3]

    start, end, name = find_section_bounds(tags_file, target_file, section_query)

    if start is None:
        print(f"Section matching '{section_query}' not found")
        sys.exit(1)

    print(f"Section: {name}")
    print(f"Lines: {start}-{end if end else 'EOF'}")
    print(f"\nTo read this section:")
    print(f"Read(file_path='{target_file}', offset={start-1}, limit={end-start+1 if end else 'None'})")


if __name__ == '__main__':
    main()

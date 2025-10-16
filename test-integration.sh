#!/bin/bash
set -e

# Integration test script for markdown-nav MCP server
# Usage: ./test-integration.sh <markdown-file>

if [ $# -eq 0 ]; then
    echo "Usage: $0 <markdown-file>"
    echo "Example: $0 mcp-mdnav-server-prompt.md"
    exit 1
fi

MARKDOWN_FILE="$1"
TAGS_FILE="${2:-tags}"
SERVER="./mdnav-server"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if server binary exists
if [ ! -f "$SERVER" ]; then
    echo -e "${RED}Error: Server binary not found at $SERVER${NC}"
    echo "Please build it first: go build -o mdnav-server"
    exit 1
fi

# Check if markdown file exists
if [ ! -f "$MARKDOWN_FILE" ]; then
    echo -e "${RED}Error: Markdown file not found: $MARKDOWN_FILE${NC}"
    exit 1
fi

# Check if tags file exists
if [ ! -f "$TAGS_FILE" ]; then
    echo -e "${YELLOW}Warning: Tags file not found: $TAGS_FILE${NC}"
    echo -e "${YELLOW}Generating tags file...${NC}"
    ctags -R --fields=+KnS --languages=markdown "$MARKDOWN_FILE"
fi

echo -e "${GREEN}=== Markdown Navigation MCP Server Integration Tests ===${NC}"
echo "Testing with file: $MARKDOWN_FILE"
echo "Using tags file: $TAGS_FILE"
echo ""

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test a tool
test_tool() {
    local tool_name="$1"
    local request="$2"
    local expected_pattern="$3"

    echo -e "${YELLOW}Testing: $tool_name${NC}"

    # Send request to server and capture response
    response=$(echo "$request" | "$SERVER" 2>/dev/null)

    # Check if response contains expected pattern
    if echo "$response" | grep -q "$expected_pattern"; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        echo "Expected pattern: $expected_pattern"
        echo "Response: $response"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Initialize MCP connection
INIT_REQUEST=$(cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}
EOF
)

echo -e "${YELLOW}Initializing MCP connection...${NC}"
init_response=$(echo "$INIT_REQUEST" | "$SERVER" 2>/dev/null | head -1)
if echo "$init_response" | grep -q '"result"'; then
    echo -e "${GREEN}✓ Connection initialized${NC}"
    echo ""
else
    echo -e "${RED}✗ Failed to initialize connection${NC}"
    echo "Response: $init_response"
    exit 1
fi

# Test 1: markdown_tree
TREE_REQUEST=$(cat <<EOF
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"markdown_tree","arguments":{"file_path":"$MARKDOWN_FILE","tags_file":"$TAGS_FILE"}}}
EOF
)
test_tool "markdown_tree" "$TREE_REQUEST" '"tree"'
echo ""

# Test 2: markdown_list_sections (all sections)
LIST_REQUEST=$(cat <<EOF
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"markdown_list_sections","arguments":{"file_path":"$MARKDOWN_FILE","tags_file":"$TAGS_FILE"}}}
EOF
)
test_tool "markdown_list_sections (all)" "$LIST_REQUEST" '"sections"'
echo ""

# Test 3: markdown_list_sections (filter by H2)
LIST_H2_REQUEST=$(cat <<EOF
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"markdown_list_sections","arguments":{"file_path":"$MARKDOWN_FILE","heading_level":"H2","tags_file":"$TAGS_FILE"}}}
EOF
)
test_tool "markdown_list_sections (H2 only)" "$LIST_H2_REQUEST" '"sections"'
echo ""

# Test 4: markdown_section_bounds (first section)
# Get first H2 section name from list
FIRST_SECTION=$(echo "$LIST_H2_REQUEST" | "$SERVER" 2>/dev/null | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$FIRST_SECTION" ]; then
    BOUNDS_REQUEST=$(cat <<EOF
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"markdown_section_bounds","arguments":{"file_path":"$MARKDOWN_FILE","section_query":"$FIRST_SECTION","tags_file":"$TAGS_FILE"}}}
EOF
)
    test_tool "markdown_section_bounds" "$BOUNDS_REQUEST" '"start_line"'
    echo ""

    # Test 5: markdown_read_section
    READ_REQUEST=$(cat <<EOF
{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"markdown_read_section","arguments":{"file_path":"$MARKDOWN_FILE","section_query":"$FIRST_SECTION","tags_file":"$TAGS_FILE"}}}
EOF
)
    test_tool "markdown_read_section" "$READ_REQUEST" '"content"'
    echo ""
else
    echo -e "${YELLOW}Skipping section-specific tests (no sections found)${NC}"
    echo ""
fi

# Print summary
echo -e "${GREEN}=== Test Summary ===${NC}"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi

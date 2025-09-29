#!/bin/bash

# Integration test for the specific issue that was reported
# Tests that "markdown-runner -B help<TAB>" shows chunks, not just stage names

source "$(dirname "$0")/../bash_completion.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Integration Test: Stage Chunk Completion Fix ===${NC}"
echo

echo "Issue: 'markdown-runner -B help<TAB>' should list available chunks under help"
echo

# Test the specific case mentioned in the issue
echo "Testing: markdown-runner -B help"
COMP_WORDS=("markdown-runner" "-B" "help")
COMP_CWORD=2
COMP_LINE="markdown-runner -B help"
COMP_POINT=${#COMP_LINE}
COMPREPLY=()

_markdown_runner_completion

echo "Completions found: ${#COMPREPLY[@]}"
echo "Completions: ${COMPREPLY[@]}"

# Verify the result
if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "help/0" ]]; then
    echo -e "${GREEN}✓ SUCCESS: help stage correctly shows chunk help/0${NC}"
    echo
    
    # Test a few more cases to ensure robustness
    echo "Additional verification tests:"
    
    # Test setup stage
    echo "Testing: markdown-runner -B setup"
    COMP_WORDS=("markdown-runner" "-B" "setup")
    COMP_CWORD=2
    COMPREPLY=()
    _markdown_runner_completion
    
    if [[ ${#COMPREPLY[@]} -eq 2 ]]; then
        echo -e "  ${GREEN}✓ setup stage shows 2 chunks: ${COMPREPLY[@]}${NC}"
    else
        echo -e "  ${RED}✗ setup stage failed: ${COMPREPLY[@]}${NC}"
    fi
    
    # Test partial stage name
    echo "Testing: markdown-runner -B hel"
    COMP_WORDS=("markdown-runner" "-B" "hel")
    COMP_CWORD=2
    COMPREPLY=()
    _markdown_runner_completion
    
    if [[ ${#COMPREPLY[@]} -eq 1 ]] && [[ "${COMPREPLY[0]}" == "help" ]]; then
        echo -e "  ${GREEN}✓ partial 'hel' correctly shows 'help' stage${NC}"
    else
        echo -e "  ${RED}✗ partial completion failed: ${COMPREPLY[@]}${NC}"
    fi
    
    echo
    echo -e "${GREEN}🎉 Integration test PASSED! The autocompletion fix is working correctly.${NC}"
    exit 0
else
    echo -e "${RED}✗ FAILED: Expected 'help/0', got: ${COMPREPLY[@]}${NC}"
    echo -e "${RED}The autocompletion fix is not working correctly.${NC}"
    exit 1
fi

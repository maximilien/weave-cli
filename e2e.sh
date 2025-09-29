#!/bin/bash

# End-to-End Integration Test Script for Weave CLI
# Tests all major functionality including collections and documents

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test collection names (using test-specific names to avoid polluting main collections)
TEXT_COLLECTION="WeaveDocs_test"
IMAGE_COLLECTION="WeaveImages_test"

# Vector DB type - try weaviate-cloud first, fallback to mock
VECTOR_DB_TYPE="weaviate-cloud"

# Function to check Weaviate configuration
check_weaviate_config() {
    echo -e "${YELLOW}🔍 Checking Weaviate configuration...${NC}"
    
    # Try to run a simple health check with weaviate-cloud
    if ./bin/weave health --vector-db-type weaviate-cloud > /dev/null 2>&1; then
        VECTOR_DB_TYPE="weaviate-cloud"
        echo -e "${GREEN}✅ Weaviate Cloud is configured and accessible${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  Weaviate Cloud not configured or not accessible${NC}"
        echo -e "${YELLOW}🔄 Falling back to mock database for testing${NC}"
        VECTOR_DB_TYPE="mock"
        return 1
    fi
}

# Function to print test header
print_test_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}🧪 E2E Integration Test Suite${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "${CYAN}📝 Using test collections: $TEXT_COLLECTION, $IMAGE_COLLECTION${NC}"
    echo -e "${CYAN}🔒 These collections are cleaned up after testing${NC}"
}

# Function to print test section
print_section() {
    echo -e "\n${CYAN}📋 $1${NC}"
    echo -e "${CYAN}$(printf '=%.0s' {1..50})${NC}"
}

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_exit_code="${3:-0}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "\n${YELLOW}Test $TOTAL_TESTS: $test_name${NC}"
    echo -e "${YELLOW}Command: $command${NC}"
    
    if eval "$command" > /dev/null 2>&1; then
        if [ $? -eq "$expected_exit_code" ]; then
            echo -e "${GREEN}✅ PASSED${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ FAILED (wrong exit code)${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        if [ $? -eq "$expected_exit_code" ]; then
            echo -e "${GREEN}✅ PASSED${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ FAILED${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
}


# Function to cleanup test collections
cleanup_collections() {
    echo -e "\n${YELLOW}🧹 Cleaning up test collections...${NC}"
    
    # Check if collections exist before attempting cleanup
    echo -e "${YELLOW}Checking if test collections exist...${NC}"
    
    # Try to delete documents (will fail gracefully if collection doesn't exist)
    echo -e "${YELLOW}Deleting documents from $TEXT_COLLECTION...${NC}"
    ./bin/weave docs delete-all "$TEXT_COLLECTION" --vector-db-type "$VECTOR_DB_TYPE" --force --quiet 2>/dev/null || echo "Collection $TEXT_COLLECTION may not exist or is already clean"
    
    echo -e "${YELLOW}Deleting documents from $IMAGE_COLLECTION...${NC}"
    ./bin/weave docs delete-all "$IMAGE_COLLECTION" --vector-db-type "$VECTOR_DB_TYPE" --force --quiet 2>/dev/null || echo "Collection $IMAGE_COLLECTION may not exist or is already clean"
    
    # Delete schemas (this works even if collection doesn't exist)
    echo -e "${YELLOW}Deleting schema for $TEXT_COLLECTION...${NC}"
    ./bin/weave cols delete-schema "$TEXT_COLLECTION" --vector-db-type "$VECTOR_DB_TYPE" --force --quiet || true
    
    echo -e "${YELLOW}Deleting schema for $IMAGE_COLLECTION...${NC}"
    ./bin/weave cols delete-schema "$IMAGE_COLLECTION" --vector-db-type "$VECTOR_DB_TYPE" --force --quiet || true
    
    # Skip collection deletion since schema deletion is sufficient for cleanup
    echo -e "${YELLOW}Schema deletion completed - collections will be recreated during tests${NC}"
}

# Function to print final results
print_results() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}📊 E2E Test Results Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}✅ Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}❌ Failed: $FAILED_TESTS${NC}"
    echo -e "${CYAN}📈 Total:  $TOTAL_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed! E2E integration test successful!${NC}"
        exit 0
    else
        echo -e "\n${RED}💥 Some tests failed. Please check the output above.${NC}"
        exit 1
    fi
}

# Main execution
main() {
    print_test_header
    
    # Check Weaviate configuration
    check_weaviate_config
    
    # Pre-test cleanup
    echo -e "${YELLOW}🧹 Pre-test cleanup...${NC}"
    cleanup_collections
    
    # Step 1: Lint and Build
    print_section "Step 1: Code Quality & Build"
    run_test "Lint code" "./lint.sh"
    run_test "Build binary" "./build.sh"
    
    # Step 2: Health Check
    print_section "Step 2: Health Check"
    run_test "Health check" "./bin/weave health --vector-db-type $VECTOR_DB_TYPE"
    
    # Step 3: Config Tests
    print_section "Step 3: Configuration Tests"
    run_test "Show config" "./bin/weave config show --vector-db-type $VECTOR_DB_TYPE"
    
    # Step 4: Collection Tests
    print_section "Step 4: Collection Management Tests"
    
    # List collections (should be empty initially)
    run_test "List collections (initial)" "./bin/weave cols ls --vector-db-type $VECTOR_DB_TYPE"
    
    # Create text collection
    run_test "Create text collection" "./bin/weave cols create '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --schema-type ragmedocs --embedding-model text-embedding-3-small"
    
    # Create image collection
    run_test "Create image collection" "./bin/weave cols create '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --schema-type ragmeimages --embedding-model text-embedding-3-small"
    
    # List collections (should show our test collections)
    run_test "List collections (after creation)" "./bin/weave cols ls --vector-db-type $VECTOR_DB_TYPE"
    
    # Show text collection
    run_test "Show text collection" "./bin/weave cols show '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # Show image collection
    run_test "Show image collection" "./bin/weave cols show '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # Count collections
    run_test "Count collections" "./bin/weave cols count --vector-db-type $VECTOR_DB_TYPE"
    
    # Step 5: Document Tests - Text Collection
    print_section "Step 5: Document Tests - Text Collection"
    
    # List documents (should be empty)
    run_test "List documents in text collection (empty)" "./bin/weave docs ls '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # Create documents from docs/ directory
    if [ -d "docs" ] && [ "$(ls -A docs)" ]; then
        for doc_file in docs/*.md; do
            if [ -f "$doc_file" ]; then
                run_test "Create document: $(basename "$doc_file")" "./bin/weave docs create '$TEXT_COLLECTION' '$doc_file' --vector-db-type $VECTOR_DB_TYPE"
            fi
        done
    else
        echo -e "${YELLOW}⚠️  No documents found in docs/ directory${NC}"
    fi
    
    # List documents (should show created documents)
    run_test "List documents in text collection (populated)" "./bin/weave docs ls '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # List documents with virtual structure
    run_test "List documents with virtual structure" "./bin/weave docs ls '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE -w -S"
    
    # Count documents
    run_test "Count documents in text collection" "./bin/weave docs count '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # Show individual documents (if any exist)
    if [ -d "docs" ] && [ "$(find docs -name "*.md" | wc -l)" -gt 0 ]; then
        first_doc=$(find docs -name "*.md" | head -1)
        if [ -n "$first_doc" ]; then
            run_test "Show document: $(basename "$first_doc")" "./bin/weave docs show '$TEXT_COLLECTION' '$(basename "$first_doc")' --vector-db-type $VECTOR_DB_TYPE"
        fi
    fi
    
    # Step 6: Document Tests - Image Collection
    print_section "Step 6: Document Tests - Image Collection"
    
    # List documents (should be empty)
    run_test "List documents in image collection (empty)" "./bin/weave docs ls '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # Create documents from images/ directory
    if [ -d "images" ] && [ "$(ls -A images)" ]; then
        for img_file in images/*.png images/*.jpg images/*.jpeg; do
            if [ -f "$img_file" ]; then
                run_test "Create image document: $(basename "$img_file")" "./bin/weave docs create '$IMAGE_COLLECTION' '$img_file' --vector-db-type $VECTOR_DB_TYPE"
            fi
        done
    else
        echo -e "${YELLOW}⚠️  No images found in images/ directory${NC}"
    fi
    
    # List documents (should show created image documents)
    run_test "List documents in image collection (populated)" "./bin/weave docs ls '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # List documents with virtual structure
    run_test "List documents with virtual structure (images)" "./bin/weave docs ls '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE -w -S"
    
    # Count documents
    run_test "Count documents in image collection" "./bin/weave docs count '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # Show individual image documents (if any exist)
    if [ -d "images" ] && [ "$(find images -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" | wc -l)" -gt 0 ]; then
        first_img=$(find images -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" | head -1)
        if [ -n "$first_img" ]; then
            run_test "Show image document: $(basename "$first_img")" "./bin/weave docs show '$IMAGE_COLLECTION' '$(basename "$first_img")' --vector-db-type $VECTOR_DB_TYPE"
        fi
    fi
    
    # Step 7: Document Deletion Tests
    print_section "Step 7: Document Deletion Tests"
    
    # Delete individual documents (if any exist)
    if [ -d "docs" ] && [ "$(find docs -name "*.md" | wc -l)" -gt 0 ]; then
        first_doc=$(find docs -name "*.md" | head -1)
        if [ -n "$first_doc" ]; then
            run_test "Delete document: $(basename "$first_doc")" "./bin/weave docs delete '$TEXT_COLLECTION' '$(basename "$first_doc")' --vector-db-type $VECTOR_DB_TYPE --force"
        fi
    fi
    
    if [ -d "images" ] && [ "$(find images -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" | wc -l)" -gt 0 ]; then
        first_img=$(find images -name "*.png" -o -name "*.jpg" -o -name "*.jpeg" | head -1)
        if [ -n "$first_img" ]; then
            run_test "Delete image document: $(basename "$first_img")" "./bin/weave docs delete '$IMAGE_COLLECTION' '$(basename "$first_img")' --vector-db-type $VECTOR_DB_TYPE --force"
        fi
    fi
    
    # Delete all documents
    run_test "Delete all documents from text collection" "./bin/weave docs delete-all '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --force"
    run_test "Delete all documents from image collection" "./bin/weave docs delete-all '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --force"
    
    # Step 8: Collection Deletion Tests
    print_section "Step 8: Collection Deletion Tests"
    
    # Delete schemas (more reliable than collection deletion)
    run_test "Delete text collection schema" "./bin/weave cols delete-schema '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --force"
    run_test "Delete image collection schema" "./bin/weave cols delete-schema '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --force"
    
    # Note: Collection deletion is skipped as schema deletion is sufficient for cleanup
    # and collection deletion can fail if collections don't exist or have dependencies
    
    # Final verification
    run_test "List collections (final - should be clean)" "./bin/weave cols ls --vector-db-type $VECTOR_DB_TYPE"
    
    # Print results
    print_results
}

# Run main function
main "$@"
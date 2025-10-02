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
    
    # Capture both stdout and stderr
    local output
    local exit_code
    
    if output=$(eval "$command" 2>&1); then
        exit_code=$?
        if [ $exit_code -eq "$expected_exit_code" ]; then
            echo -e "${GREEN}✅ PASSED${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ FAILED (wrong exit code: $exit_code, expected: $expected_exit_code)${NC}"
            echo -e "${RED}Output: $output${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        exit_code=$?
        if [ $exit_code -eq "$expected_exit_code" ]; then
            echo -e "${GREEN}✅ PASSED${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ FAILED (exit code: $exit_code)${NC}"
            echo -e "${RED}Output: $output${NC}"
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

    # Step 3a: Schema Configuration Tests
    print_section "Step 3a: Schema Configuration Tests"
    run_test "List configured schemas" "./bin/weave config list-schemas"
    run_test "Show RagMeDocs schema" "./bin/weave config show-schema RagMeDocs"
    run_test "Show RagMeImages schema" "./bin/weave config show-schema RagMeImages"
    run_test "Show TestSchema from directory" "./bin/weave config show-schema TestSchema"

    # Test error handling for non-existent schema
    echo -e "${YELLOW}Testing error handling for non-existent schema...${NC}"
    if ! ./bin/weave config show-schema NonExistentSchema > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Correctly returned error for non-existent schema${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ Should have errored for non-existent schema${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Step 4: Collection Tests
    print_section "Step 4: Collection Management Tests"
    
    # List collections (should be empty initially)
    run_test "List collections (initial)" "./bin/weave cols ls --vector-db-type $VECTOR_DB_TYPE"
    
    # Create text collection using WeaveDocs schema with flat metadata
    run_test "Create text collection" "./bin/weave collection create '$TEXT_COLLECTION' --text --flat-metadata --vector-db-type $VECTOR_DB_TYPE"

    # Create image collection using WeaveImages schema with flat metadata
    run_test "Create image collection" "./bin/weave collection create '$IMAGE_COLLECTION' --image --flat-metadata --vector-db-type $VECTOR_DB_TYPE"

    # Test creating collection from named schema in directory
    TEST_SCHEMA_COLLECTION="TestSchemaCollection"
    run_test "Create collection from directory schema" "./bin/weave cols create '$TEST_SCHEMA_COLLECTION' --schema TestSchema --vector-db-type $VECTOR_DB_TYPE"

    # List collections (should show our test collections)
    run_test "List collections (after creation)" "./bin/weave cols ls --vector-db-type $VECTOR_DB_TYPE"
    
    # Show text collection
    run_test "Show text collection" "./bin/weave cols show '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # Show image collection
    run_test "Show image collection" "./bin/weave cols show '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE"
    
    # Show collection schema
    run_test "Show collection schema" "./bin/weave cols show '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --schema"
    
    # Show collection metadata
    run_test "Show collection metadata" "./bin/weave cols show '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --metadata"
    
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
    
    # Show expanded metadata analysis
    run_test "Show expanded metadata analysis" "./bin/weave cols show '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --expand-metadata"
    
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
    
    # Create documents from test images directory (small images for reliable testing)
    if [ -d "tests/images" ] && [ "$(ls -A tests/images)" ]; then
        for img_file in tests/images/*.png tests/images/*.jpg tests/images/*.jpeg; do
            if [ -f "$img_file" ]; then
                run_test "Create image document: $(basename "$img_file")" "./bin/weave docs create '$IMAGE_COLLECTION' '$img_file' --vector-db-type $VECTOR_DB_TYPE"
            fi
        done
    else
        echo -e "${YELLOW}⚠️  No test images found in tests/images/ directory${NC}"
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

    # Step 6.5: Query Tests (before documents are deleted)
    print_section "Step 6.5: Query Tests (with documents)"
    
    # Test basic semantic search
    run_test "Basic semantic search" "./bin/weave cols q '$TEXT_COLLECTION' 'weave-cli installation' --vector-db-type $VECTOR_DB_TYPE"
    
    # Test search with custom result limit
    run_test "Search with custom result limit" "./bin/weave cols q '$TEXT_COLLECTION' 'machine learning' --top_k 3 --vector-db-type $VECTOR_DB_TYPE"
    
    # Test search with metadata flag
    run_test "Search with metadata flag" "./bin/weave cols q '$TEXT_COLLECTION' 'README' --search-metadata --vector-db-type $VECTOR_DB_TYPE"
    
    # Test case insensitive search
    run_test "Case insensitive search" "./bin/weave cols q '$TEXT_COLLECTION' 'WEAVE-CLI' --vector-db-type $VECTOR_DB_TYPE"
    
    # Test BM25 search
    run_test "BM25 keyword search" "./bin/weave cols q '$TEXT_COLLECTION' 'sample' --bm25 --vector-db-type $VECTOR_DB_TYPE"
    
    # Test query help
    run_test "Query help command" "./bin/weave cols q --help"

    # Step 6.6: Compact Schema Tests (before documents are deleted)
    print_section "Step 6.6: Compact Schema Tests (with documents)"

    # Test compact flag - should show metadata but remove occurrences/samples
    run_test "Export schema with compact flag to YAML" "./bin/weave cols show '$TEXT_COLLECTION' --schema --yaml-file /tmp/${TEXT_COLLECTION}_schema_compact.yaml --compact --vector-db-type $VECTOR_DB_TYPE"
    run_test "Export schema with compact flag to JSON" "./bin/weave cols show '$TEXT_COLLECTION' --schema --json-file /tmp/${TEXT_COLLECTION}_schema_compact.json --compact --vector-db-type $VECTOR_DB_TYPE"
    run_test "Display compact schema as YAML" "./bin/weave cols show '$TEXT_COLLECTION' --schema --yaml --compact --vector-db-type $VECTOR_DB_TYPE"

    # Verify compact output has metadata fields but no occurrences/samples, and no empty nestedproperties
    echo -e "${YELLOW}Verifying compact mode...${NC}"
    # Check for metadata fields (either single metadata field or individual metadata properties)
    if grep -q "metadata:" /tmp/${TEXT_COLLECTION}_schema_compact.yaml || grep -q "name: added_date" /tmp/${TEXT_COLLECTION}_schema_compact.yaml || grep -q "name: creation_date" /tmp/${TEXT_COLLECTION}_schema_compact.yaml || grep -q "name: filename" /tmp/${TEXT_COLLECTION}_schema_compact.yaml; then
        has_metadata=true
    else
        has_metadata=false
    fi
    
    if ! grep -q "occurrences:" /tmp/${TEXT_COLLECTION}_schema_compact.yaml; then
        no_occurrences=true
    else
        no_occurrences=false
    fi
    
    if ! grep -q "sample:" /tmp/${TEXT_COLLECTION}_schema_compact.yaml; then
        no_samples=true
    else
        no_samples=false
    fi
    
    if ! grep -q "nestedproperties: \[\]" /tmp/${TEXT_COLLECTION}_schema_compact.yaml; then
        no_empty_nested=true
    else
        no_empty_nested=false
    fi
    
    if [ "$has_metadata" = "true" ] && [ "$no_occurrences" = "true" ] && [ "$no_samples" = "true" ] && [ "$no_empty_nested" = "true" ]; then
        echo -e "${GREEN}✅ Compact mode verified: Has metadata fields but no occurrences/samples or empty nested properties${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ Compact mode failed: Missing metadata fields or has occurrences/samples/empty nested properties${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

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

    # Step 7.5: Schema Export and Import Tests
    print_section "Step 7.5: Schema Export and Import Round-trip Tests"

    # Export schema to YAML
    run_test "Export text collection schema to YAML" "./bin/weave cols show '$TEXT_COLLECTION' --schema --yaml-file /tmp/${TEXT_COLLECTION}_schema.yaml --vector-db-type $VECTOR_DB_TYPE"

    # Export schema to JSON
    run_test "Export text collection schema to JSON" "./bin/weave cols show '$TEXT_COLLECTION' --schema --json-file /tmp/${TEXT_COLLECTION}_schema.json --vector-db-type $VECTOR_DB_TYPE"

    # Display YAML schema to stdout
    run_test "Display schema as YAML" "./bin/weave cols show '$TEXT_COLLECTION' --schema --yaml --vector-db-type $VECTOR_DB_TYPE"

    # Display JSON schema to stdout
    run_test "Display schema as JSON" "./bin/weave cols show '$TEXT_COLLECTION' --schema --json --vector-db-type $VECTOR_DB_TYPE"

    # Create new collection from exported schema
    SCHEMA_TEST_COLLECTION="${TEXT_COLLECTION}_SchemaRoundTrip"
    run_test "Create collection from YAML schema file" "./bin/weave cols create '$SCHEMA_TEST_COLLECTION' --schema-yaml-file /tmp/${TEXT_COLLECTION}_schema.yaml --vector-db-type $VECTOR_DB_TYPE"

    # Export schema from newly created collection
    run_test "Export schema from round-trip collection" "./bin/weave cols show '$SCHEMA_TEST_COLLECTION' --schema --yaml-file /tmp/${SCHEMA_TEST_COLLECTION}_schema.yaml --vector-db-type $VECTOR_DB_TYPE"

    # Verify schemas match (ignoring collection name and metadata occurrences)
    echo -e "${YELLOW}Verifying schema round-trip integrity...${NC}"
    if diff <(grep -A 50 "^schema:" /tmp/${TEXT_COLLECTION}_schema.yaml | grep -v "occurrences" | sed "s/${TEXT_COLLECTION}/${SCHEMA_TEST_COLLECTION}/g") <(grep -A 50 "^schema:" /tmp/${SCHEMA_TEST_COLLECTION}_schema.yaml | grep -v "occurrences") > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Schema round-trip verified: Schemas match!${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ Schema round-trip failed: Schemas do not match${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Delete the schema test collection
    run_test "Delete schema round-trip test collection" "./bin/weave cols delete-schema '$SCHEMA_TEST_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --force"

    # Step 8: Collection Deletion Tests
    print_section "Step 8: Collection Deletion Tests"
    
    # Delete schemas (more reliable than collection deletion)
    run_test "Delete text collection schema" "./bin/weave cols delete-schema '$TEXT_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --force"
    run_test "Delete image collection schema" "./bin/weave cols delete-schema '$IMAGE_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --force"
    run_test "Delete test schema collection" "./bin/weave cols delete-schema '$TEST_SCHEMA_COLLECTION' --vector-db-type $VECTOR_DB_TYPE --force"

    # Note: Collection deletion is skipped as schema deletion is sufficient for cleanup
    # and collection deletion can fail if collections don't exist or have dependencies

    # Step 9: Pattern Matching Tests
    print_section "Step 9: Pattern Matching Tests"
    
    # Create test collections with different patterns
    PATTERN_TEST_COLLECTIONS=("PatternTest1" "PatternTest2" "TestPattern3" "OtherCollection")
    for collection in "${PATTERN_TEST_COLLECTIONS[@]}"; do
        run_test "Create pattern test collection: $collection" "./bin/weave cols create '$collection' --text --flat-metadata --vector-db-type $VECTOR_DB_TYPE"
    done
    
    # Test glob pattern matching for collection deletion
    run_test "Delete collections matching glob pattern 'PatternTest*'" "./bin/weave cols delete-schema --pattern 'PatternTest*' --vector-db-type $VECTOR_DB_TYPE --force"
    
    # Test regex pattern matching for collection deletion
    run_test "Delete collections matching regex pattern 'Test.*'" "./bin/weave cols delete-schema --pattern 'Test.*' --vector-db-type $VECTOR_DB_TYPE --force"
    
    # Verify remaining collections
    run_test "List collections after pattern deletion" "./bin/weave cols ls --vector-db-type $VECTOR_DB_TYPE"
    
    # Clean up remaining test collections
    run_test "Delete remaining test collection" "./bin/weave cols delete-schema 'OtherCollection' --vector-db-type $VECTOR_DB_TYPE --force"

    # Final verification
    run_test "List collections (final - should be clean)" "./bin/weave cols ls --vector-db-type $VECTOR_DB_TYPE"
    
    # Print results
    print_results
}

# Run main function
main "$@"
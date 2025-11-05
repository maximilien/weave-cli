# MCP Integration Tests

## Overview

The `mcp_integration_test.go` file contains automated integration tests for
verifying that weave-cli works correctly with the latest weave-mcp server
release.

## Prerequisites

1. **Environment variables** - Create a `.env` file in the project root with:

   ```bash
   OPENAI_API_KEY=sk-proj-your-key
   WEAVE_MCP_STDIO_PATH=/path/to/weave-mcp/bin/weave-mcp-stdio
   WEAVIATE_URL=https://your-cluster.weaviate.cloud
   WEAVIATE_API_KEY=your-weaviate-key
   ```

2. **weave-mcp server** - Ensure the weave-mcp binary is built and accessible
   at the path specified in `WEAVE_MCP_STDIO_PATH`

## Running the Tests

### Run all MCP integration tests

```bash
go test -v -tags=integration ./tests -run TestMCP -timeout 5m
```

### Run specific test

```bash
go test -v -tags=integration ./tests -run TestMCPIntegration -timeout 5m
```

### Run with environment variables directly

```bash
OPENAI_API_KEY=sk-proj-xxx \
WEAVE_MCP_STDIO_PATH=/path/to/weave-mcp-stdio \
WEAVIATE_URL=https://your-cluster.weaviate.cloud \
WEAVIATE_API_KEY=xxx \
go test -v -tags=integration ./tests -run TestMCP -timeout 5m
```

## Test Coverage

The MCP integration tests verify:

1. **Collection Operations**
   - Create text and image collections
   - List all collections
   - Delete collections

2. **Document Operations**
   - Count documents in collections
   - List documents in collections

3. **System Operations**
   - Health checks
   - MCP tool schema validation

4. **Error Handling**
   - Duplicate collection creation
   - Non-existent collection operations

## Automated Testing for MCP Releases

When a new weave-mcp release is made:

1. Update the MCP binary:

   ```bash
   cd /path/to/weave-mcp
   npm run build
   ```

2. Run the integration tests:

   ```bash
   cd /path/to/weave-cli
   go test -v -tags=integration ./tests -run TestMCP -timeout 5m
   ```

3. Verify all tests pass before considering the release compatible

## Troubleshooting

### Test fails with "EOF" error

- Check that `WEAVE_MCP_STDIO_PATH` points to a valid binary
- Ensure the MCP binary has execute permissions:
  `chmod +x /path/to/weave-mcp-stdio`
- Verify all required environment variables are set

### Test fails with "OPENAI_API_KEY not set"

- Ensure `.env` file exists in project root
- Verify the `.env` file contains `OPENAI_API_KEY=...`
- Or set the environment variable directly before running tests

### Test hangs or times out

- Increase timeout: `-timeout 10m`
- Check if MCP server is responding: Test manually with `weave query` commands
- Enable verbose logging: Set `Verbose: true` in test config

## CI/CD Integration

To run these tests in CI/CD:

```yaml
- name: Run MCP Integration Tests
  env:
    OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
    WEAVE_MCP_STDIO_PATH: ./weave-mcp/bin/weave-mcp-stdio
    WEAVIATE_URL: ${{ secrets.WEAVIATE_URL }}
    WEAVIATE_API_KEY: ${{ secrets.WEAVIATE_API_KEY }}
  run: |
    go test -v -tags=integration ./tests -run TestMCP -timeout 5m
```

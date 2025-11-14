# Improved Error Messages

The weave CLI now features simplified, human-friendly error messages that help you quickly understand and fix issues.

## Before and After Examples

### Collection Creation Errors

#### Before (verbose and repetitive):
```
❌ Failed to create collection: failed to create collection: failed to create collection 'WeaveDocs': failed to create collection: HTTP 422 - {"error":[{"message":"module 'text2vec-openai': wrong OpenAI model name, available model names are: [ada babbage curie davinci text-embedding-3-small text-embedding-3-large]"}]}
```

#### After (clean and actionable):
```
❌ Failed to create collection: Invalid OpenAI embedding model. Available models:
  • ada
  • babbage
  • curie
  • davinci
  • text-embedding-3-small
  • text-embedding-3-large

Use --embedding-model flag to specify a valid model.
```

### Document Creation Errors

#### Before:
```
❌ Failed to create Weaviate document: failed to create document: failed to insert document: collection not found: HTTP 404 - collection 'MyDocs' does not exist
```

#### After:
```
❌ Failed to create document: collection 'MyDocs' does not exist
```

### Collection Deletion Errors

#### Before:
```
❌ Failed to delete collection(s): failed to delete collection: failed to delete: connection refused: dial tcp: connection refused
```

#### After:
```
❌ Failed to delete collection(s): connection refused
```

## How It Works

The error simplification system:

1. **Extracts JSON error messages** from Weaviate API responses
2. **Removes redundant prefixes** (e.g., multiple "failed to create collection")
3. **Formats specific errors** with helpful suggestions:
   - OpenAI model errors → List of valid models
   - Missing collections → Clear message about collection not existing
   - Connection errors → Simple connectivity message
4. **Removes HTTP status codes** that don't add value to the user

## Features

- **Context-aware formatting**: Different formatting for creation, deletion, and query errors
- **Actionable suggestions**: Tells you what flag to use or what to check
- **Clean output**: Removes technical jargon and HTTP noise
- **Preserves important details**: Keeps collection names, model names, and other relevant info

## Implementation

All error formatting is handled by the `utils.SimplifyError()` function and related helpers in `src/cmd/utils/errors.go`.

Commands automatically use the appropriate formatter:
- `FormatCreationError()` for create operations
- `FormatDeletionError()` for delete operations
- `FormatQueryError()` for query operations

## Testing

Error formatting is fully tested in `src/cmd/utils/errors_test.go` with test cases covering:
- Simple errors
- Redundant prefixes
- Weaviate JSON errors
- OpenAI model errors
- Collection not found errors
- Connection errors

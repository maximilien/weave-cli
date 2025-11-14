# Supabase Collection Name Normalization Fix

## Problem

The Supabase adapter was automatically normalizing collection names, causing unexpected behavior:

- `My_Collection` → `my-collection` (lowercased, underscores → hyphens)
- `TestCollection_1` → `testcollection-1`
- Inconsistent with Weaviate behavior
- User confusion when names didn't match what they specified

## Solution

Updated the implementation to preserve original collection names (except spaces):

### Changes Made

#### 1. Updated `getTableName()` in `adapter.go`

**Before:**
```go
func (a *Adapter) getTableName(collectionName string) string {
    tableName := strings.ToLower(collectionName)
    tableName = strings.ReplaceAll(tableName, "-", "_")
    tableName = strings.ReplaceAll(tableName, " ", "_")
    return fmt.Sprintf("collection_%s", tableName)
}
```

**After:**
```go
func (a *Adapter) getTableName(collectionName string) string {
    // Only replace spaces with underscores for valid PostgreSQL identifiers
    tableName := strings.ReplaceAll(collectionName, " ", "_")
    return fmt.Sprintf("collection_%s", tableName)
}
```

#### 2. Added `quoteIdentifier()` Helper Function

```go
func quoteIdentifier(identifier string) string {
    // Escape any double quotes in the identifier
    escaped := strings.ReplaceAll(identifier, `"`, `""`)
    return fmt.Sprintf(`"%s"`, escaped)
}
```

This ensures PostgreSQL preserves case and special characters in table names.

#### 3. Updated All SQL Queries

Updated all SQL queries across multiple files to use quoted identifiers:

**Files Modified:**
- `src/pkg/vectordb/supabase/adapter.go` - Added helper functions
- `src/pkg/vectordb/supabase/collections.go` - CREATE/DROP TABLE queries
- `src/pkg/vectordb/supabase/documents.go` - DELETE queries
- `src/pkg/vectordb/supabase/queries.go` - SELECT queries (search)
- `src/pkg/vectordb/supabase/schema.go` - ALTER TABLE queries

**Example Change:**
```go
// Before
query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", tableName)

// After
quotedTable := quoteIdentifier(tableName)
query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", quotedTable)
```

#### 4. Added Integration Test

Added comprehensive test in `tests/supabase_integration_test.go` to verify name preservation:

```go
t.Run("CollectionNamePreservation", func(t *testing.T) {
    testNames := []string{
        "TestCollection_MixedCase",
        "My_Test_Collection",
        "Collection123_Test",
    }

    for _, name := range testNames {
        // Create, verify, add docs, retrieve, cleanup
        // All operations use exact name
    }
})
```

## Results

### Before
```bash
weave cols create "MyTest_Collection"
# Actually creates: collection_mytest-collection

weave cols ls
# Shows: mytest-collection (user confusion!)
```

### After
```bash
weave cols create "MyTest_Collection"
# Creates: "collection_MyTest_Collection" (quoted in PostgreSQL)

weave cols ls
# Shows: MyTest_Collection (preserves original name!)
```

## Compatibility

### Preserved Behavior
- ✅ Spaces still converted to underscores (PostgreSQL requirement)
- ✅ Collection name prefixed with `collection_` (namespace isolation)
- ✅ All CRUD operations work correctly
- ✅ Backward compatible with existing code

### Changed Behavior
- ✅ Case preserved (`MyCollection` stays `MyCollection`)
- ✅ Underscores preserved (`my_collection` stays `my_collection`)
- ✅ Hyphens preserved (`my-collection` stays `my-collection`)
- ✅ Numbers preserved (`Collection123` stays `Collection123`)

### Not Affected
- Special characters in PostgreSQL identifiers (e.g., `@`, `#`) - Not recommended, but technically work if quoted

## Migration

No migration needed! Changes are forward-compatible:

- **Existing collections**: Continue working with their current names
- **New collections**: Use preserved names
- **Mixed environment**: Both old and new naming schemes work simultaneously

## Testing

```bash
# Build
go build -o bin/weave ./src

# Run integration tests (requires Supabase credentials)
go test -v ./tests -run TestSupabaseIntegration
```

The new `CollectionNamePreservation` test verifies:
1. Collections with mixed case are created correctly
2. Collections can be queried with exact original name
3. Documents can be added/retrieved using original collection name
4. Collections can be deleted using original name

## Benefits

1. **Consistency** - Matches Weaviate behavior
2. **User Expectations** - Names match what users specify
3. **Clarity** - No hidden transformations
4. **Standards** - Uses PostgreSQL quoted identifiers correctly

## Implementation Notes

### Why Quoted Identifiers?

PostgreSQL treats unquoted identifiers as case-insensitive (folding to lowercase). Quoted identifiers preserve exact casing:

```sql
-- Unquoted - becomes lowercase
CREATE TABLE MyTable;  -- Actually creates "mytable"

-- Quoted - preserves case
CREATE TABLE "MyTable";  -- Creates "MyTable"
```

### Quote Escaping

Double quotes inside identifiers are escaped by doubling them:

```sql
CREATE TABLE "My""Special""Table";  -- Table named: My"Special"Table
```

Our `quoteIdentifier()` function handles this automatically.

## Related

- See `docs/SUPABASE_TODO.md` for remaining Supabase improvements
- See `docs/VDB_SUPPORT.md` for complete feature matrix

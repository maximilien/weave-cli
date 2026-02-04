# Config Fix Feature Implementation Plan

**Created**: 2026-02-04
**Status**: Planning
**Target Version**: v0.9.16

---

## Overview

Add `weave config fix` command to help users interactively fix configuration errors detected by validation.

**Current State**:
- ✅ Excellent error reporting with hints (see config_errors.go)
- ✅ Validation shows errors/warnings with field paths
- ✅ Interactive config update exists (`weave config update`)
- ❌ No automated fix workflow for validation errors

**Desired State**:
- ✅ `weave config fix` command that reads validation errors
- ✅ Interactive prompts for each error/warning
- ✅ Smart defaults and suggestions
- ✅ Dry-run mode to preview changes
- ✅ Batch fix mode for automation

---

## Example User Flow

### Current Experience (Without Fix)
```bash
➜ weave config update --weave-mcp

⚠️  Configuration Errors:
  • databases.vector_databases[2] (mongodb-cloud).database_url: MongoDB database_url is required
    💡 Set 'database_url' field with MongoDB Atlas connection string

⚠️  Configuration Warnings:
  • databases.vector_databases[4] (milvus-cloud).api_key: Milvus Cloud (Zilliz) requires an API key
    💡 Set 'api_key' field for authentication
  • databases.vector_databases[7] (chroma-cloud).tenant: Chroma Cloud requires a tenant ID
    💡 Set 'tenant' field to your Chroma Cloud tenant ID

# User now has to manually edit config.yaml
```

### Proposed Experience (With Fix)
```bash
➜ weave config fix

🔍 Found 1 error and 2 warnings in config.yaml

Would you like to fix these issues? (Y/n): y

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Error 1/1: mongodb-cloud missing database_url
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Database: mongodb-cloud (databases.vector_databases[2])
Field: database_url
Issue: MongoDB database_url is required

💡 Tip: Use your MongoDB Atlas connection string
   Example: mongodb+srv://username:password@cluster.mongodb.net/database

Options:
  1. Enter value now
  2. Skip this database (disable it)
  3. Remove this database from config
  4. Quit

Your choice (1-4): 1

Enter database_url: mongodb+srv://myuser:***@prod.mongodb.net/weave
✅ Value set

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Warning 1/2: milvus-cloud missing api_key
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Database: milvus-cloud (databases.vector_databases[4])
Field: api_key
Issue: Milvus Cloud (Zilliz) requires an API key

💡 Tip: Get your API key from https://cloud.zilliz.com/settings/api-keys

Options:
  1. Enter value now
  2. Skip (fix later)
  3. Remove this database from config
  4. Quit

Your choice (1-4): 2
⏭️  Skipped

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Warning 2/2: chroma-cloud missing tenant
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Database: chroma-cloud (databases.vector_databases[7])
Field: tenant
Issue: Chroma Cloud requires a tenant ID

💡 Tip: Find your tenant ID at https://cloud.trychroma.com/settings

Options:
  1. Enter value now
  2. Skip (fix later)
  3. Remove this database from config
  4. Quit

Your choice (1-4): 1

Enter tenant: my-tenant-123
✅ Value set

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Changes to be made:
  ✅ mongodb-cloud.database_url = mongodb+srv://myuser:***@prod.mongodb.net/weave
  ⏭️  milvus-cloud.api_key (skipped)
  ✅ chroma-cloud.tenant = my-tenant-123

Apply changes to config.yaml? (Y/n): y

✅ Configuration updated successfully!

Run 'weave config update --weave-mcp' to validate.
```

---

## Command Design

### Command Signature
```bash
weave config fix [flags]
```

### Flags
```bash
--config-file string     Config file to fix (default: "config.yaml")
--dry-run               Show what would be fixed without making changes
--auto-fix              Automatically fix warnings (skip prompts for warnings)
--errors-only           Only fix errors, ignore warnings
--non-interactive       Use in scripts, fail if user input needed
--backup                Create backup before making changes (default: true)
--format string         Output format: text, json (default: "text")
```

### Usage Examples
```bash
# Interactive fix (default)
weave config fix

# Dry run - see what would be fixed
weave config fix --dry-run

# Auto-fix warnings, prompt for errors
weave config fix --auto-fix

# Fix only errors
weave config fix --errors-only

# Non-interactive (for CI/CD)
weave config fix --non-interactive --errors-only

# Custom config file
weave config fix --config-file ~/.weave-cli/config.yaml

# JSON output for scripting
weave config fix --format json
```

---

## Technical Design

### 1. Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    weave config fix                          │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ├─► Load config.yaml
                     │
                     ├─► Run validation (existing code)
                     │
                     ├─► Parse errors/warnings
                     │   └─► ConfigIssue struct
                     │       • Type: error/warning
                     │       • Path: field path
                     │       • Database: name
                     │       • Field: field name
                     │       • Message: description
                     │       • Hint: fix suggestion
                     │
                     ├─► Interactive fix loop
                     │   └─► For each issue:
                     │       • Display context
                     │       • Show options
                     │       • Prompt for action
                     │       • Collect fixes
                     │
                     ├─► Show summary
                     │
                     ├─► Apply fixes to config
                     │   └─► YAML manipulation
                     │       • Preserve comments
                     │       • Preserve formatting
                     │       • Atomic write
                     │
                     └─► Validate again
```

### 2. New Data Structures

```go
// ConfigIssue represents a validation issue with fix information
type ConfigIssue struct {
    Type        IssueType  // error, warning
    Path        string     // e.g., "databases.vector_databases[2].database_url"
    Database    string     // e.g., "mongodb-cloud"
    Field       string     // e.g., "database_url"
    Message     string     // Description of issue
    Hint        string     // Fix suggestion
    FixOptions  []FixOption
}

type IssueType string

const (
    IssueTypeError   IssueType = "error"
    IssueTypeWarning IssueType = "warning"
)

// FixOption represents a way to fix an issue
type FixOption struct {
    ID          string  // e.g., "enter_value", "skip", "remove", "disable"
    Label       string  // e.g., "Enter value now"
    Description string  // Optional longer description
    Action      FixAction
}

type FixAction func(issue ConfigIssue, config *Config) error

// FixResult tracks the outcome of fixing an issue
type FixResult struct {
    Issue       ConfigIssue
    Action      string  // "fixed", "skipped", "removed", "disabled"
    Value       string  // New value (masked if secret)
    Error       error
}
```

### 3. Key Functions

```go
// ParseValidationErrors converts validation output to ConfigIssue list
func ParseValidationErrors(validationOutput string) ([]ConfigIssue, error)

// InteractiveFix prompts user to fix each issue
func InteractiveFix(issues []ConfigIssue, config *Config) ([]FixResult, error)

// ApplyFixes applies fixes to config file atomically
func ApplyFixes(configPath string, fixes []FixResult) error

// ValidateAfterFix re-runs validation to confirm fixes
func ValidateAfterFix(configPath string) error
```

### 4. YAML Manipulation

**Challenge**: Need to modify YAML while preserving:
- Comments
- Formatting
- Field order
- Indentation

**Solution**: Use `gopkg.in/yaml.v3` with node-based editing:

```go
import "gopkg.in/yaml.v3"

// SetNestedValue sets a value in YAML while preserving structure
func SetNestedValue(node *yaml.Node, path string, value string) error {
    // Path: "databases.vector_databases[2].database_url"
    // Navigate to node
    // Update value
    // Preserve comments and style
}

// RemoveNestedValue removes a field or array element
func RemoveNestedValue(node *yaml.Node, path string) error

// DisableDatabase comments out a database config
func DisableDatabase(node *yaml.Node, index int) error
```

---

## Implementation Plan

### Phase 1: Core Infrastructure (2-3 hours)

**Files**:
- `src/cmd/config/fix.go` (new)
- `src/pkg/config/fix.go` (new)
- `src/pkg/config/yaml_edit.go` (new)

**Tasks**:
1. Create `ConfigIssue` and related structs
2. Implement `ParseValidationErrors()`
3. Create basic command structure
4. Add YAML editing utilities

**Deliverables**:
- Basic command that parses validation errors
- YAML editing foundation

### Phase 2: Interactive UI (3-4 hours)

**Files**:
- `src/pkg/config/fix_interactive.go` (new)

**Tasks**:
1. Design interactive prompt flow
2. Implement issue display with context
3. Add options menu (enter/skip/remove/disable)
4. Implement input collection (with secret masking)
5. Add summary display

**Deliverables**:
- Full interactive experience
- Secret handling

### Phase 3: Fix Application (2-3 hours)

**Files**:
- `src/pkg/config/fix_apply.go` (new)

**Tasks**:
1. Implement atomic config file updates
2. Add backup functionality
3. Handle different fix actions (set/remove/disable)
4. Preserve YAML structure

**Deliverables**:
- Safe config file updates
- Automatic backups

### Phase 4: Testing & Polish (2-3 hours)

**Files**:
- `src/cmd/config/fix_test.go` (new)
- `src/pkg/config/fix_test.go` (new)

**Tasks**:
1. Unit tests for parsing
2. Unit tests for YAML editing
3. Integration tests
4. Add `--dry-run` and `--auto-fix` flags
5. Error handling and edge cases

**Deliverables**:
- Comprehensive test coverage
- Robust error handling

### Phase 5: Documentation (1 hour)

**Files**:
- `docs/USER_GUIDE.md` (update)
- `README.md` (update)
- `CHANGELOG.md` (update)

**Tasks**:
1. Add usage examples
2. Document all flags
3. Add troubleshooting section

---

## Edge Cases & Considerations

### 1. Multiple Errors in Same Database
**Issue**: If `mongodb-cloud` has 3 missing fields
**Solution**: Group by database, fix all fields for that DB together

### 2. Array Index Changes
**Issue**: Removing `vector_databases[2]` changes indices of later items
**Solution**: Process removals after all updates, in reverse index order

### 3. Environment Variable References
**Issue**: `url: ${WEAVIATE_URL}` - do we edit .env or config?
**Solution**:
- Detect `${VAR}` syntax
- Offer to edit .env file instead
- Show: "This field uses ${VAR}. Would you like to update .env file?"

### 4. Secret Leakage
**Issue**: API keys in terminal history
**Solution**:
- Always use `term.ReadPassword()` for secrets
- Mask in all output (e.g., `sk-proj-abc...xyz`)
- Never log secrets

### 5. Concurrent Edits
**Issue**: User edits config.yaml while fix is running
**Solution**:
- Check file modification time before writing
- Warn if file changed: "config.yaml was modified. Continue? (y/n)"

### 6. Invalid Fix Values
**Issue**: User enters invalid URL/API key format
**Solution**:
- Add validation for common fields
- Re-prompt on validation failure
- Show format examples

---

## Smart Fix Suggestions

### Auto-Detect Patterns

**1. Missing URL**
```
Detected: field named "url", "database_url", "host"
Suggestion: Check for WEAVIATE_URL in .env
Prompt: "Found WEAVIATE_URL=https://... in .env. Use this? (Y/n)"
```

**2. Missing API Key**
```
Detected: field named "*api_key", "*token", "*secret"
Suggestion: Check for matching env vars
Prompt: "Found OPENAI_API_KEY in .env. Use this? (Y/n)"
```

**3. Cloud Database Not Used**
```
Detected: Warning on cloud DB with missing credentials
Suggestion: Remove if not needed
Prompt: "Are you using mongodb-cloud? (y/N)"
If no: "Remove from config? (Y/n)"
```

---

## Non-Interactive Mode

For CI/CD and automation:

```bash
# Exit with error if fixes needed
weave config fix --non-interactive --errors-only

# Exit codes:
# 0 = No issues or all fixed
# 1 = Errors remain (user input needed)
# 2 = Warnings remain (with --errors-only)
```

**JSON Output** (for scripting):
```json
{
  "status": "needs_fixes",
  "errors": 1,
  "warnings": 2,
  "issues": [
    {
      "type": "error",
      "path": "databases.vector_databases[2].database_url",
      "database": "mongodb-cloud",
      "field": "database_url",
      "message": "MongoDB database_url is required",
      "hint": "Set 'database_url' field with MongoDB Atlas connection string"
    }
  ]
}
```

---

## Future Enhancements (v0.9.17+)

1. **AI-Powered Fixes**:
   - Use Claude to suggest fixes based on context
   - Example: "Based on your other configs, should this be weaviate-cloud or milvus-cloud?"

2. **Config Templates**:
   - `weave config fix --template=production`
   - Pre-fill with production-ready defaults

3. **Bulk Operations**:
   - `weave config fix --disable-all-warnings`
   - `weave config fix --remove-unused-databases`

4. **Schema Validation**:
   - Validate URLs are reachable
   - Check API keys work (with user consent)
   - Verify database connections

5. **Config Migration**:
   - `weave config fix --migrate-from=0.9.12`
   - Auto-update config format between versions

---

## Success Metrics

**Usability**:
- ✅ User can fix all errors in <2 minutes
- ✅ Clear, actionable prompts
- ✅ No accidental data loss

**Reliability**:
- ✅ 100% of valid fixes succeed
- ✅ Config file always valid YAML after fix
- ✅ Automatic backups prevent data loss

**Coverage**:
- ✅ Handles all validation errors
- ✅ Handles all validation warnings
- ✅ Works with nested configs

---

## Testing Plan

### Unit Tests
```go
TestParseValidationErrors()
TestSetNestedValue()
TestRemoveNestedValue()
TestDisableDatabase()
TestApplyFixes()
```

### Integration Tests
```go
TestFixMissingDatabaseURL()
TestFixMultipleIssues()
TestFixWithBackup()
TestDryRun()
TestNonInteractive()
```

### Manual Tests
- [ ] Fix error in mongodb-cloud
- [ ] Fix multiple warnings
- [ ] Skip some issues
- [ ] Remove database from config
- [ ] Disable database
- [ ] Dry run mode
- [ ] Auto-fix mode
- [ ] Non-interactive mode fails correctly
- [ ] Backup created and valid
- [ ] YAML comments preserved

---

## Total Effort Estimate

**Development**: 10-13 hours
**Testing**: 3-4 hours
**Documentation**: 1 hour

**Total**: 14-18 hours (~2-3 days)

---

## Dependencies

**New**:
- `gopkg.in/yaml.v3` (for node-based YAML editing)

**Existing**:
- `github.com/fatih/color` (already used)
- `golang.org/x/term` (already used)
- `github.com/spf13/cobra` (already used)

---

## References

- Current validation: `src/cmd/utils/config_helper.go`
- Error formatting: `src/pkg/config/config_errors.go`
- Interactive config: `src/cmd/config/update.go`
- YAML v3 docs: https://pkg.go.dev/gopkg.in/yaml.v3

---

## Decision Log

**2026-02-04**: Initial design
- Decided on interactive-first approach with non-interactive fallback
- Chose gopkg.in/yaml.v3 for structure-preserving edits
- Prioritized user safety (backups, validation, confirmation)

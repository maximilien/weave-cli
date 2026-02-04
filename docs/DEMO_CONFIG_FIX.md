# Weave Config Fix - Demo

## Overview

`weave config fix` is an interactive tool to help users fix configuration errors and warnings.

## Features

✅ **Interactive Prompts**: Step-by-step guidance for each issue
✅ **Secret Masking**: Hides API keys/passwords during input  
✅ **Multiple Fix Options**: Enter value, skip, remove database, disable database, quit
✅ **Automatic Backups**: Creates timestamped backups before changes
✅ **Dry Run Mode**: Preview changes without applying
✅ **Smart Validation**: Detects missing required fields by database type
✅ **Re-validation**: Confirms fixes worked after applying

## Usage

### Basic Interactive Fix

```bash
weave config fix
```

Detects issues like:
- ❌ **Error**: mongodb-cloud.database_url missing
- ⚠️  **Warning**: milvus-cloud.api_key missing  
- ⚠️  **Warning**: chroma-cloud.tenant missing

### Dry Run (Preview Only)

```bash
weave config fix --dry-run
```

Shows what would be fixed without making changes.

### Fix Errors Only

```bash
weave config fix --errors-only
```

Ignores warnings, only fixes errors.

### Custom Config File

```bash
weave config fix --config-file ~/.weave-cli/config.yaml
```

## Interactive Flow Example

```
🔍 Validating configuration...

Found 1 error and 2 warnings
   • 1 error(s)
   • 2 warning(s)

Would you like to fix these issues? (Y/n): y

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Error 1/1: mongodb-cloud missing database_url
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Database: mongodb-cloud (databases.vector_databases[2])
Field: database_url
Issue: MongoDB database_url is required

💡 Tip: Set 'database_url' field with MongoDB Atlas connection string
   Example: mongodb+srv://user:pass@cluster.mongodb.net/db

Options:
  1. Enter value now
  2. Skip (fix later)
  3. Remove this database from config
  4. Disable this database (comment out)
  5. Quit

Your choice (1-5): 1

Enter database_url: mongodb+srv://user:pass@prod.mongodb.net/weave
✅ Value set

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Warning 1/2: milvus-cloud missing api_key
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Database: milvus-cloud (databases.vector_databases[4])
Field: api_key
Issue: Milvus Cloud (Zilliz) requires an API key

💡 Tip: Set 'api_key' field for authentication
   Example: your-api-key-here

Options:
  1. Enter value now
  2. Skip (fix later)
  3. Remove this database from config
  4. Disable this database (comment out)
  5. Quit

Your choice (1-5): 1

Enter api_key (input hidden): 
✅ Value set

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Warning 2/2: chroma-cloud missing tenant
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Database: chroma-cloud (databases.vector_databases[7])
Field: tenant
Issue: Chroma Cloud requires a tenant ID

💡 Tip: Set 'tenant' field to your Chroma Cloud tenant ID
   Example: your-tenant-id

Options:
  1. Enter value now
  2. Skip (fix later)
  3. Remove this database from config
  4. Disable this database (comment out)
  5. Quit

Your choice (1-5): 3
🗑️  Will remove database from config

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Changes to be made:
  ✅ mongodb-cloud.database_url = mongodb+srv://user:***@prod.mongodb.net/weave
  ✅ milvus-cloud.api_key = sk-p...xyz
  🗑️  Remove database: chroma-cloud

Apply changes to config.yaml? (Y/n): y

💾 Applying changes...
💾 Backup created: config.backup-20260204-130045.yaml

✅ Configuration updated successfully!

🔍 Re-validating configuration...
✅ Configuration is now valid!

Next steps:
  • Test your configuration: weave config show
  • List databases: weave config list
  • Try it out: weave cols ls --weaviate-cloud
```

## Fix Options Explained

### 1. Enter value now
- Prompts for the value
- Masks secrets (API keys show as `sk-...xyz`)
- Validates non-empty input

### 2. Skip (fix later)
- Leaves the issue for next time
- Useful when you don't have the value ready
- No changes made to config

### 3. Remove this database from config
- Completely deletes the database entry
- Good for unused cloud services
- Index updated automatically

### 4. Disable this database (comment out)
- Comments out the database config
- Adds `enabled: false` field
- Preserves config for later re-enabling

### 5. Quit
- Stops without applying any changes
- Safe exit at any time
- No backup created

## Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--config-file` | Custom config file | `--config-file ~/.weave-cli/config.yaml` |
| `--dry-run` | Preview without applying | `weave config fix --dry-run` |
| `--auto-fix` | Auto-fix warnings | `weave config fix --auto-fix` |
| `--errors-only` | Only fix errors | `weave config fix --errors-only` |
| `--no-backup` | Skip backup creation | `weave config fix --no-backup` |

## Validation Rules

### MongoDB
- ❌ `database_url` required
- Format: `mongodb+srv://user:pass@cluster.mongodb.net/db`

### Milvus Cloud (Zilliz)
- ⚠️  `api_key` recommended
- Get from: https://cloud.zilliz.com/settings/api-keys

### Chroma Cloud
- ⚠️  `tenant` recommended
- Find at: https://cloud.trychroma.com/settings

### Qdrant Cloud
- ⚠️  `api_key` recommended
- Get from: https://cloud.qdrant.io/settings/api-keys

## Backup Management

Backups are created automatically with timestamp:

```bash
config.yaml                          # Original
config.backup-20260204-130045.yaml  # Auto backup
```

To restore:
```bash
cp config.backup-20260204-130045.yaml config.yaml
```

## Testing & Development

### Unit Tests (All Passing ✅)

```bash
go test -v ./src/pkg/config/fix_test.go ./src/pkg/config/fix.go
go test -v ./src/pkg/config/yaml_edit_test.go ./src/pkg/config/yaml_edit.go
```

### Build

```bash
./build.sh
./bin/weave config fix --help
```

### Manual Testing

```bash
# Requires actual terminal (not script)
./bin/weave config fix
```

## Implementation Details

- **Language**: Go
- **Dependencies**: 
  - `gopkg.in/yaml.v3` - Structure-preserving YAML editing
  - `github.com/fatih/color` - Colored terminal output
  - `golang.org/x/term` - Secret input masking
- **Files**:
  - `src/pkg/config/fix.go` - Core parsing
  - `src/pkg/config/fix_interactive.go` - Interactive UI
  - `src/pkg/config/fix_apply.go` - Fix application
  - `src/pkg/config/yaml_edit.go` - YAML manipulation
  - `src/cmd/config/fix.go` - CLI command

## Known Limitations

1. **Terminal Required**: Interactive mode needs a real terminal (won't work in CI/CD scripts)
2. **Limited Validation**: Currently only checks required fields, not formats/connectivity
3. **Single File**: Only works with config.yaml, not .env files

## Future Enhancements

- [ ] Auto-fix mode implementation
- [ ] Environment variable detection and prompts
- [ ] URL/API key format validation
- [ ] Connection testing after fixes
- [ ] Config migration between versions
- [ ] Support for .env file fixes
- [ ] Non-interactive mode with defaults

## Related Commands

- `weave config show` - View current configuration
- `weave config list` - List configured databases
- `weave config update --env` - Update .env file
- `weave config create --config-yaml` - Create new config

## Documentation

- [USER_GUIDE.md](USER_GUIDE.md) - Complete usage guide
- [CONFIG_FIX_FEATURE.md](planning/CONFIG_FIX_FEATURE.md) - Implementation plan
- [OBSERVABILITY.md](OBSERVABILITY.md) - Production features

---

**Weave CLI** - Vector Database Orchestration Tool  
**Version**: v0.9.16+ (config fix feature)

# Weave CLI Demo Scripts

Interactive demonstration scripts for Weave CLI features.

## Available Demos

### 🔧 Configuration Demo

**File**: `config-demo.sh`

**Duration**: ~5 minutes

**Description**: Interactive walkthrough of Weave CLI configuration management.

**Topics**:

- Viewing current configuration
- Creating `.env` file interactively with `weave config create --env`
- Environment variable configuration
- Global vs local configuration
- Configuration precedence (CLI flags → env vars → config.yaml → defaults)
- Installing weave-mcp for REPL mode
- Health check verification

**Run**:

```bash
./demos/config-demo.sh
```

**Prerequisites**:

- None! Works with any configuration or mock database

---

### 🗄️ Supabase Integration Demo

**File**: `supabase-demo.sh`

**Duration**: ~8 minutes

**Description**: Comprehensive demonstration of Supabase (PostgreSQL +
pgvector) integration.

**Topics**:

- Supabase configuration setup
- Creating collections in Supabase
- Adding documents
- Semantic search with pgvector
- BM25 full-text keyword search
- Hybrid search (combining vector + BM25)
- Multi-database operations (querying both Weaviate and Supabase)
- Supabase-specific features and limitations

**Run**:

```bash
./demos/supabase-demo.sh
```

**Prerequisites**:

- Supabase project with pgvector extension enabled
- Environment variables configured:

  ```bash
  export SUPABASE_DATABASE_URL="postgres://postgres:[password]@db.[project].supabase.co:6543/postgres"
  export SUPABASE_DATABASE_KEY="your-anon-key"
  export VECTOR_DB_TYPE="supabase"
  ```

**Setup Supabase**:

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Verify extension
SELECT * FROM pg_extension WHERE extname = 'vector';
```

---

### ⚡ Quick Demo

**File**: `quick-demo.sh`

**Duration**: ~2 minutes

**Description**: Fast overview of core Weave CLI functionality.

**Topics**:

- Health check
- List collections
- Create collection
- Add document
- List documents
- Query collection

**Run**:

```bash
./demos/quick-demo.sh
```

**Prerequisites**:

- Weaviate or mock database configured

---

### 📚 Full Demo

**File**: `full-demo.sh`

**Duration**: ~5 minutes

**Description**: Complete walkthrough of all Weave CLI features.

**Topics**:

- Configuration management
- Health checks
- Collection creation and management
- Document operations
- Semantic search
- Available embeddings
- Multi-database support
- Batch processing overview

**Run**:

```bash
./demos/full-demo.sh
```

**Prerequisites**:

- Weaviate or mock database configured

---

### 🤖 REPL Demo

**File**: `repl-demo.sh`

**Duration**: ~4 minutes

**Description**: AI-powered natural language interface demonstration.

**Topics**:

- REPL introduction
- Prerequisites check
- Natural language query examples
- Multi-agent system
- Cost tracking
- Dry-run mode
- Interactive REPL mode
- Opik observability integration

**Run**:

```bash
./demos/repl-demo.sh
```

**Prerequisites**:

- OPENAI_API_KEY environment variable
- weave-mcp binary (install with `weave config update --weave-mcp`)

---

## Video Demos

Pre-recorded demos available on asciinema:

- **[Full Demo (5 min)](https://asciinema.org/a/LrKzmThBfDbTPISZzr8biP4dt)**
  \- Complete feature walkthrough
- **[Quick Demo (2 min)](https://asciinema.org/a/HiAU7h1iJvZ2QdJe70ae3Cc0b)**
  \- Quick overview
- **[REPL Demo](https://asciinema.org/a/U504HN4FSeMsOA0qS0os0NWUE)** -
  AI-powered interface

---

## Automated Recording

Record all demos automatically with one command:

```bash
./tools/record-all-demos.sh
```

Options:

1. Config Demo (always available)
2. Supabase Demo (requires Supabase configured)
3. Full Demo (5 min - requires Weaviate configured)
4. Quick Demo (2 min - requires Weaviate configured)
5. REPL Demo (requires OpenAI API key and weave-mcp)
6. All Available Demos

The script will:

- ✅ Check prerequisites automatically
- ✅ Record demos using asciinema
- ✅ Offer to upload to asciinema.org
- ✅ Track upload URLs in `videos/latest-demo-uploads.txt`
- ✅ Show how to review and upload recordings

For more details, see [tools/README-DEMO-RECORDING.md](../tools/README-DEMO-RECORDING.md)

---

## Creating Your Own Demo

### Method 1: Using asciinema Tool

```bash
# Run the interactive recording tool
./tools/asciinema.sh

# Choose demo type:
# 1. Full Demo (5 min)
# 2. Quick Demo (2 min)
# 3. REPL Demo
# 4. Custom Demo
```

### Method 2: Manual Recording

```bash
# Install asciinema (if not installed)
brew install asciinema  # macOS
# or
apt-get install asciinema  # Linux

# Start recording
asciinema rec my-demo.cast

# Run your demo commands
./demos/config-demo.sh

# Stop recording (Ctrl+D or exit)
exit

# Upload to asciinema.org (optional)
asciinema upload my-demo.cast
```

---

## Demo Script Guidelines

When creating new demo scripts:

1. **Clear Structure**: Use numbered pages/sections
2. **Interactive**: Pause between sections with `read -p "Press Enter..."`
3. **Colored Output**: Use color codes for better readability
4. **Error Handling**: Use `|| true` or `|| echo` for optional commands
5. **Cleanup**: Offer cleanup at the end
6. **Documentation**: Include clear comments and help text

**Template**:

```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WEAVE_BIN="$PROJECT_ROOT/bin/weave"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "================================================"
echo "  Demo Title"
echo "================================================"

# Page 1
echo -e "${BLUE}📋 Page 1: Introduction${NC}"
echo "Description..."
read -p "Press Enter to continue..."
clear

# ... more pages ...
```

---

## Testing Demos

Before recording or publishing:

1. **Test locally**: Run the script multiple times
2. **Check timing**: Ensure each section has appropriate pauses
3. **Verify output**: Check that all commands produce expected results
4. **Handle errors**: Test with missing prerequisites
5. **Clean state**: Start with clean environment (no leftover test data)

---

## Contributing

To add a new demo:

1. Create script in `demos/` directory
2. Follow naming convention: `feature-demo.sh`
3. Make executable: `chmod +x demos/feature-demo.sh`
4. Update `demos/README.md` (this file)
5. Update `docs/guides/DEMO.md`
6. Test thoroughly
7. Submit PR

---

## Documentation

- **Full Demo Guide**: [docs/guides/DEMO.md](../docs/guides/DEMO.md)
- **User Guide**: [docs/USER_GUIDE.md](../docs/USER_GUIDE.md)
- **Supabase Docs**: [docs/supabase/README.md](../docs/supabase/README.md)
- **Main README**: [README.md](../README.md)

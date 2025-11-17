# Automated Demo Recording

Quick guide for recording all Weave CLI demos automatically.

## Quick Start

```bash
# Install expect for full automation (optional but recommended)
brew install expect  # macOS
# or
apt-get install expect  # Linux

# Run the automated recording script
./tools/record-all-demos.sh
```

## What It Does

The `record-all-demos.sh` script:

1. **Checks prerequisites** - Verifies asciinema, weave binary, and configuration
2. **Offers options** - Select which demos to record
3. **Records automatically** - Uses `expect` to automate pressing Enter through demos
4. **Uploads to asciinema.org** - Optionally uploads recordings
5. **Tracks URLs** - Saves upload URLs to `videos/latest-demo-uploads.txt`

## Demo Options

1. **Config Demo** - Always available (no external dependencies)
2. **Supabase Demo** - Requires Supabase configured
3. **Full Demo** - Requires Weaviate (manual recording)
4. **Quick Demo** - Requires Weaviate (manual recording)
5. **All Available Demos** - Records all configured demos

## Requirements

### Essential

- **asciinema** - For recording terminal sessions

  ```bash
  brew install asciinema  # macOS
  apt-get install asciinema  # Linux
  ```

- **weave binary** - Built and ready

  ```bash
  ./build.sh
  ```

### Recommended

- **expect** - For fully automated recording (auto-presses Enter)

  ```bash
  brew install expect  # macOS
  apt-get install expect  # Linux
  ```

  Without expect, you'll need to manually press Enter through the demos.

### Configuration

Depending on which demos you want to record:

**For Config Demo**:

- No configuration needed! Works with any setup.

**For Supabase Demo**:

```bash
export SUPABASE_DATABASE_URL="postgres://postgres:[password]@db.[project].supabase.co:6543/postgres"
export SUPABASE_DATABASE_KEY="your-anon-key"
export VECTOR_DB_TYPE="supabase"
```

**For Full/Quick Demos**:

```bash
export VECTOR_DB_TYPE="weaviate-cloud"
export WEAVIATE_URL="https://your-instance.weaviate.cloud"
export WEAVIATE_API_KEY="your-api-key"
export OPENAI_API_KEY="sk-proj-your-key"
```

Or use mock database:

```bash
export VECTOR_DB_TYPE="mock"
```

## Usage Examples

### Record Single Demo

```bash
./tools/record-all-demos.sh
# Select option 1 for Config Demo
```

### Record All Available Demos

```bash
./tools/record-all-demos.sh
# Select option 5 for All Available Demos
```

### Manual Recording (for Full/Quick demos)

Full and Quick demos require following scripts in `docs/guides/DEMO.md`:

```bash
./tools/record-all-demos.sh
# Select option 3 or 4
# Follow the demo script manually
# asciinema records your terminal session
```

## Output

Recordings are saved to:

```text
videos/
├── weave-cli-config-demo.cast
├── weave-cli-supabase-demo.cast
├── weave-cli-full-demo.cast
└── weave-cli-quick-demo.cast
```

Upload URLs are tracked in:

```text
videos/latest-demo-uploads.txt
```

## Workflow

1. **Record demos**:

   ```bash
   ./tools/record-all-demos.sh
   ```

2. **Review recordings** (optional):

   ```bash
   asciinema play videos/weave-cli-config-demo.cast
   ```

3. **Upload to asciinema.org**:
   - Script offers to upload automatically
   - Or manually: `asciinema upload videos/weave-cli-config-demo.cast`

4. **Update documentation**:
   - Copy URLs from `videos/latest-demo-uploads.txt`
   - Update `docs/guides/DEMO.md`
   - Update `README.md`

5. **Commit changes**:

   ```bash
   git add videos/ docs/ README.md
   git commit -m "chore: update demo recordings"
   git push
   ```

## Advanced Usage

### Record Specific Demo Manually

```bash
# Start recording
asciinema rec videos/weave-cli-config-demo.cast

# Run demo with expect automation
./tools/auto-demo-recorder.exp ./demos/config-demo.sh

# Or run demo manually
./demos/config-demo.sh

# Stop recording (Ctrl+D if running manually)
```

### Customize Recording

Edit demo scripts in `demos/`:

- `config-demo.sh` - Configuration demo
- `supabase-demo.sh` - Supabase integration demo

Edit `tools/auto-demo-recorder.exp` to customize automation behavior.

## Troubleshooting

### "asciinema not found"

Install asciinema:

```bash
brew install asciinema  # macOS
apt-get install asciinema  # Linux
```

### "expect not found"

The script works without expect, but you'll need to press Enter manually:

```bash
brew install expect  # macOS (recommended)
```

Or continue in manual mode when prompted.

### "Weaviate not configured"

For Full/Quick demos:

```bash
# Option 1: Use Weaviate
export VECTOR_DB_TYPE="weaviate-cloud"
export WEAVIATE_URL="..."
export WEAVIATE_API_KEY="..."

# Option 2: Use mock database
export VECTOR_DB_TYPE="mock"
```

### "Supabase not configured"

For Supabase demo:

```bash
export SUPABASE_DATABASE_URL="postgres://..."
export SUPABASE_DATABASE_KEY="..."
export VECTOR_DB_TYPE="supabase"

# Verify
./bin/weave health check
```

## Files

- **`tools/record-all-demos.sh`** - Main recording script (interactive menu)
- **`tools/auto-demo-recorder.exp`** - Expect script for automation
- **`demos/config-demo.sh`** - Config demo script
- **`demos/supabase-demo.sh`** - Supabase demo script
- **`docs/guides/DEMO.md`** - Full demo script (manual)
- **`videos/RECORD_ALL_DEMOS.sh`** - Quick reference guide (informational)

## See Also

- **Demo Update Plan**: `docs/planning/DEMO_UPDATE_PLAN.md`
- **Demo Documentation**: `docs/guides/DEMO.md`
- **Demo Scripts README**: `demos/README.md`

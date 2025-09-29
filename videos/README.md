# Weave CLI Demo Videos

This directory contains asciinema recordings of Weave CLI demonstrations.

## Available Recordings

- `weave-cli-full-demo.cast` - Complete 5-minute demo showcasing all features
- `weave-cli-quick-demo.cast` - Quick 2-minute demo for rapid overview

## Creating Recordings

Use the asciinema recording tool:

```bash
# Record full demo (5 minutes)
./tools/asciinema.sh demo

# Record quick demo (2 minutes)
./tools/asciinema.sh quick

# List available recordings
./tools/asciinema.sh list

# Upload to asciinema.org
./tools/asciinema.sh upload
```

## Prerequisites

1. **Install asciinema**:

   ```bash
   ./tools/asciinema.sh install
   ```

2. **Configure Weaviate**: Ensure your Weaviate Cloud instance is configured

3. **Demo Data**: Ensure you have:
   - `docs/README.md` and other markdown files
   - `images/` directory with sample images (optional)

## Playing Recordings

```bash
# Play a recording locally
asciinema play videos/weave-cli-full-demo.cast

# Upload and get shareable URL
asciinema upload videos/weave-cli-full-demo.cast
```

## Demo Script

The demo follows the script in `docs/DEMO.md` with 10 pages:

1. **Health Check & Configuration** - Verify setup
2. **Create Collections** - Text and image collections
3. **List Collections** - Show available collections
4. **Create Documents** - Add sample documents
5. **Show Documents & Schema** - Detailed document view
6. **List Documents** - Simple and virtual document listing
7. **Delete Documents** - Document deletion with confirmation
8. **Cleanup Operations** - Delete all and schema cleanup
9. **Getting Weave CLI** - Download and build instructions
10. **Thank You** - Credits and references

## Recording Tips

- **Timing**: Each command has appropriate delays for readability
- **Error Handling**: Commands include fallbacks for missing files
- **Cleanup**: Demo collections are cleaned up automatically
- **Duration**: Full demo ~5 minutes, quick demo ~2 minutes

## Sharing

Once uploaded to asciinema.org, you'll get a shareable URL like:
`https://asciinema.org/a/abc123`

Perfect for:

- Documentation
- Presentations
- Social media
- GitHub README
- Project showcases

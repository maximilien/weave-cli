#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Quick reference for recording all demos
# This is a guide - run commands manually to review each demo

set -e

cat << 'EOF'
┌─────────────────────────────────────────────────┐
│  Weave CLI - Demo Recording Quick Reference    │
└─────────────────────────────────────────────────┘

This guide walks you through recording all demos.
Run each command manually to review before recording.

Prerequisites:
  • asciinema installed (brew install asciinema)
  • Weaviate configured (for full/quick demos)
  • Supabase configured (for Supabase demo)
  • Clean environment (no test collections)

═══════════════════════════════════════════════════
 Priority 1: New Demos
═══════════════════════════════════════════════════

1️⃣  CONFIG DEMO (~5 min)
───────────────────────────────────────────────────
Prerequisites: None (works with any setup)

Test the script first:
  ./demos/config-demo.sh

Record:
  asciinema rec videos/weave-cli-config-demo.cast
  ./demos/config-demo.sh
  (Ctrl+D when done)

Upload:
  asciinema upload videos/weave-cli-config-demo.cast

Save URL to: videos/latest-demo-uploads.txt
Update: docs/guides/DEMO.md, README.md

═══════════════════════════════════════════════════

2️⃣  SUPABASE DEMO (~8 min)
───────────────────────────────────────────────────
Prerequisites: Supabase configured

Setup:
  export SUPABASE_DATABASE_URL="postgres://..."
  export SUPABASE_DATABASE_KEY="..."
  export VECTOR_DB_TYPE="supabase"
  ./bin/weave health check

Test the script first:
  ./demos/supabase-demo.sh

Record:
  asciinema rec videos/weave-cli-supabase-demo.cast
  ./demos/supabase-demo.sh
  (Ctrl+D when done)

Upload:
  asciinema upload videos/weave-cli-supabase-demo.cast

Save URL to: videos/latest-demo-uploads.txt
Update: docs/guides/DEMO.md, README.md

═══════════════════════════════════════════════════
 Priority 2: Update Existing Demos
═══════════════════════════════════════════════════

3️⃣  FULL DEMO (~5 min)
───────────────────────────────────────────────────
Prerequisites: Weaviate configured

Current script: docs/guides/DEMO.md (Pages 1-11)

Suggested additions:
  • Page 1: Add "weave embeddings list"
  • Page 2: Show "--embedding" flag usage
  • Page 7: Add "--json" flag example
  • Page 10: Reference config/Supabase demos

After updating docs/guides/DEMO.md:
  asciinema rec videos/weave-cli-full-demo.cast
  (Follow updated DEMO.md script manually)
  (Ctrl+D when done)

Upload:
  asciinema upload videos/weave-cli-full-demo.cast

Save URL to: videos/latest-demo-uploads.txt
Update: docs/guides/DEMO.md, README.md

═══════════════════════════════════════════════════

4️⃣  QUICK DEMO (~2 min)
───────────────────────────────────────────────────
Prerequisites: Weaviate configured

Current: videos/weave-cli-quick-demo.cast
Consider creating: docs/guides/QUICK_DEMO.md script

Suggested content:
  • Quick config check
  • Create collection with embedding
  • Add document
  • Query with JSON output
  • Links to other demos

After creating/updating script:
  asciinema rec videos/weave-cli-quick-demo.cast
  (Follow script)
  (Ctrl+D when done)

Upload:
  asciinema upload videos/weave-cli-quick-demo.cast

Save URL to: videos/latest-demo-uploads.txt
Update: docs/guides/DEMO.md, README.md

═══════════════════════════════════════════════════
 REPL Demo - Already Current ✓
═══════════════════════════════════════════════════

✅ REPL demo was recorded on Nov 5, 2025
   Includes latest REPL progress improvements
   URL: https://asciinema.org/a/U504HN4FSeMsOA0qS0os0NWUE
   No update needed

═══════════════════════════════════════════════════
 After Recording All Demos
═══════════════════════════════════════════════════

1. Update videos/latest-demo-uploads.txt with all URLs
2. Update docs/guides/DEMO.md:
   - Move demos from "To Be Recorded" to "Available"
   - Add new URLs
3. Update README.md with new demo links
4. Verify all links work
5. Commit changes:
   git add videos/ docs/ README.md demos/
   git commit -m "chore: update all demo recordings with latest features"
   git push

═══════════════════════════════════════════════════
 Testing Recordings
═══════════════════════════════════════════════════

Before uploading, review each recording:
  asciinema play videos/weave-cli-{demo-name}.cast

Check for:
  • Clear output
  • No errors
  • Proper timing (not too fast/slow)
  • Complete demonstration

═══════════════════════════════════════════════════

See docs/planning/DEMO_UPDATE_PLAN.md for detailed information.

EOF

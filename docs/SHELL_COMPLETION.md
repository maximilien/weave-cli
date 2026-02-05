# Shell Completion for Weave CLI

Weave CLI supports shell completion for Bash, Zsh, Fish, and PowerShell. This allows you to press `Tab` to:
- Complete command names
- Complete subcommand names
- Complete flag names
- Complete database names from your config
- Complete collection names (when connected)

## Quick Setup

### Bash

**One-time setup:**
```bash
# Add to your ~/.bashrc or ~/.bash_profile
echo 'source <(weave completion bash)' >> ~/.bashrc

# Or for global installation (requires sudo)
weave completion bash | sudo tee /etc/bash_completion.d/weave > /dev/null
```

**Then reload:**
```bash
source ~/.bashrc
```

### Zsh

**One-time setup:**
```bash
# Add to your ~/.zshrc
echo 'source <(weave completion zsh)' >> ~/.zshrc

# Or for global installation
weave completion zsh | sudo tee /usr/local/share/zsh/site-functions/_weave > /dev/null
```

**Then reload:**
```bash
source ~/.zshrc
```

### Fish

**One-time setup:**
```bash
weave completion fish | source

# Or to make it persistent
weave completion fish > ~/.config/fish/completions/weave.fish
```

### PowerShell

**One-time setup:**
```powershell
weave completion powershell | Out-String | Invoke-Expression

# Or to make it persistent, add to your PowerShell profile
weave completion powershell >> $PROFILE
```

## What Gets Completed

### Commands & Subcommands
```bash
weave co<TAB>       # → completes to: weave config
weave config s<TAB> # → completes to: weave config show
weave cols l<TAB>   # → completes to: weave cols list
```

### Flags
```bash
weave --ve<TAB>     # → completes to: weave --verbose
weave --vdb<TAB>    # → completes to: weave --vdb
```

### Database Types
```bash
weave --vector-db-type we<TAB>  # → completes to: weave --vector-db-type weaviate-cloud
weave --vdb mi<TAB>             # → completes to: weave --vdb milvus-local
```

### Dynamic Completions (Future)
Coming soon:
- Database names from your config.yaml
- Collection names when connected to a database
- Document IDs for delete/update operations
- MCP server names

## Verification

Test your completion setup:
```bash
weave co<TAB>              # Should complete 'config'
weave config <TAB>         # Should show: create, show, update, validate, fix
weave --v<TAB>             # Should show: --verbose, --vdb, --vector-db-type
```

## Troubleshooting

**Bash completion not working?**
- Make sure `bash-completion` is installed: `brew install bash-completion` (macOS) or `apt install bash-completion` (Linux)
- Restart your terminal or run `source ~/.bashrc`

**Zsh completion not working?**
- Make sure completion system is enabled in `~/.zshrc`:
  ```bash
  autoload -Uz compinit
  compinit
  ```
- Run `rm -f ~/.zcompdump && compinit` to rebuild completion cache

**Fish completion not working?**
- Check file was created: `ls ~/.config/fish/completions/weave.fish`
- Restart fish shell

## Advanced: Dynamic Completions

Weave CLI can provide dynamic completions based on your configuration. To enable:

```bash
# Coming soon - this will complete database names from config.yaml
weave cols list --vdb <TAB>  # → shows: weaviate-cloud, mongodb-cloud, milvus-local, etc.

# Coming soon - this will complete collection names from connected DB
weave docs list <TAB>         # → shows: products, customers, documents, etc.
```

## Tips

- Use `--help` with tab completion: `weave <TAB>` then add `--help` to any command
- Combine with aliases: Add to your shell RC file:
  ```bash
  alias w='weave'
  alias wc='weave config'
  alias wcl='weave cols list'
  ```
  Then use: `w c<TAB>` → `w config`

## Documentation

For more information:
- Run `weave completion --help` to see all available shells
- Run `weave completion bash --help` for Bash-specific instructions
- Visit: https://github.com/maximilien/weave-cli/docs/SHELL_COMPLETION.md

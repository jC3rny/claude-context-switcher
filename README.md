# cx - Claude Code context switcher

A minimal CLI tool for switching between multiple Claude Code accounts on macOS. Stores OAuth tokens in the native Keychain, right next to the original `Claude Code-credentials` entry.

## Why

Claude Code stores a single OAuth token in macOS Keychain under `Claude Code-credentials`. If you work with multiple accounts (personal, team, enterprise), logging in with one overwrites the other. `cx` solves this by saving each token as a named context in Keychain.

## How it works

- Tokens are stored in macOS Keychain as `Claude Code-credentials (<name>)`
- `cx use <name>` launches `claude` with `ANTHROPIC_AUTH_TOKEN` set, bypassing the default keychain lookup
- Active context is tracked in `~/.claude/.active-context`
- No tokens are stored in plain text files

## Install

```sh
go install github.com/jC3rny/claude-context-switcher/cmd/cx@latest
```

Or build from source:

```sh
git clone https://github.com/jC3rny/claude-context-switcher.git
cd claude-context-switcher
make build
mv cx /usr/local/bin/
```

## Development

```sh
make lint       # run gofmt + go vet + staticcheck + gosec
make build      # build cx binary
make clean      # remove binaries
make            # lint + build
```

## Usage

```
cx <command> [args]
```

### Commands

| Command | Description |
|---|---|
| `cx list` | List all saved contexts |
| `cx save <name>` | Save current keychain token as a named context |
| `cx login <name>` | Open Claude for login, then save token as context |
| `cx use <name>` | Launch Claude Code with a saved context |
| `cx delete <name>` | Delete a saved context |
| `cx show <name>` | Show token preview (first/last 6 chars) |
| `cx current` | Show the currently active context |

### Workflow

```sh
# Save your current login
cx save personal

# Login with another account and save it
cx login work
# -> opens Claude, run /login, then exit

# Switch between accounts
cx use personal
cx use work

# Check what's saved
cx list
```

### Keychain entries

After saving contexts, your Keychain will contain:

```
Claude Code-credentials              <- original (managed by Claude Code)
Claude Code-credentials (personal)   <- saved by cx
Claude Code-credentials (work)       <- saved by cx
```

## Security

- Tokens are stored in macOS Keychain (encrypted), never in plain text files
- Context names are validated: alphanumeric, hyphens, dots, underscores only
- Security scanned with [gosec](https://github.com/securego/gosec) in CI
- `cx use` passes the token via `ANTHROPIC_AUTH_TOKEN` env var (same mechanism Claude Code uses internally)

## Requirements

- macOS (uses `security` CLI for Keychain access)
- Go 1.21+ (to build)
- Claude Code CLI (`claude`) in PATH

## License

GPLv3 - see [LICENSE](LICENSE)

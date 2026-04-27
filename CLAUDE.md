# cx - Claude Code context switcher

## Project overview

A Go CLI tool (`cx`) that manages multiple Claude Code OAuth tokens in macOS Keychain. It allows switching between different Claude accounts without repeated /login cycles.

## Architecture

- Single-file Go application (`cmd/cx/main.go`), no external dependencies
- Shell completions embedded via `//go:embed` from `cmd/cx/completions/`
- Uses macOS `security` CLI to interact with Keychain
- Keychain entries follow the pattern `Claude Code-credentials (<context-name>)`
- Active context tracked in `~/.claude/.active-context`
- `cx use` replaces the process via `syscall.Exec` (no wrapper overhead)

## Key constants

- Keychain service: `Claude Code-credentials`
- Context service pattern: `Claude Code-credentials (<name>)`
- Active file: `~/.claude/.active-context`
- Account: current OS username

## Commands

- `list` - parses `security dump-keychain` output to find all cx entries
- `save` - copies token from `Claude Code-credentials` to a named entry (does NOT update the active context)
- `login` - launches `claude` interactively, saves the resulting token, and sets it as the active context
- `use` - reads named token, sets `ANTHROPIC_AUTH_TOKEN`, execs `claude`
- `delete` - removes named keychain entry
- `show` - displays token preview (first/last 6 chars + length)
- `current` - reads `~/.claude/.active-context`
- `completion` - prints embedded shell completion script to stdout; pipe or redirect as needed (e.g. `eval "$(cx completion zsh)"`)
- `version`, `--version` - prints build version (injected via ldflags)

## Flags

- `-v`, `--verbose` - verbose output to stderr (`[verbose]` prefix)
- `--debug` - debug output to stderr (`[debug]` prefix, includes verbose)
- Flags must appear before the command: `cx -v list`

## Conventions

- No external Go dependencies (stdlib only)
- macOS-only (Keychain dependency)
- Exit codes: 0 = success, 1 = error
- Error messages go to stderr, output to stdout
- Binary name: `cx`
- Context names validated against `^[a-zA-Z0-9][a-zA-Z0-9._-]*$`
- Version injected at build time via `-ldflags -X main.version=`

## Security

- golangci-lint runs gosec, govet, staticcheck, and gofmt in a single pass
- `#nosec G204` annotations are used for intentional `exec.Command` patterns where args are validated or constant
- All errors must be handled — unhandled errors will fail gosec
- govulncheck scans for known vulnerabilities in dependencies

## CI/CD

- `.github/workflows/ci.yml` - runs on push/PR to main: golangci-lint, shellcheck, govulncheck, unit tests, build (amd64 + arm64)
- `.github/workflows/release.yml` - runs on `v*` tags: same quality checks, builds stripped binaries, generates checksums and changelog, creates GitHub Release
- Build targets: `darwin/amd64` and `darwin/arm64` only
- Release binaries use `-ldflags="-s -w"` for size reduction

## Makefile targets

- `make lint` - golangci-lint (govet, staticcheck, gosec, gofmt)
- `make shellcheck` - shellcheck on bash completion script
- `make vulncheck` - govulncheck for known vulnerabilities
- `make test` - unit tests (`go test`)
- `make build` - build cx binary
- `make clean` - remove binaries
- `make` - all of the above (default)

## Testing changes

After modifying `cmd/cx/main.go`, verify:

```sh
make
./cx list && ./cx current
```

Do not run `cx save` or `cx delete` with real context names during testing. Use a disposable name like `test-tmp` and clean up after.

## Definition of done

When you believe development is complete, do NOT commit or tell the user you're done. Instead, ask the user to confirm you should run the finish procedure. The finish procedure is:

1. **Code quality and security scans** — run `make` (golangci-lint, shellcheck, govulncheck, unit tests, build)
2. **Functional tests** — run `./cx list && ./cx current && ./cx version` and test any changed/new commands
3. **Documentation** — update README.md and CLAUDE.md to reflect all changes
4. **Re-run checks** — run `make` again after any doc or code touch-ups

Only after all steps pass cleanly, report the results and ask the user if they want to commit.

# cx - Claude Code context switcher

## Project overview

A Go CLI tool (`cx`) that manages multiple Claude Code OAuth tokens in macOS Keychain. It allows switching between different Claude accounts without repeated /login cycles.

## Architecture

- Single-file Go application (`cmd/cx/main.go`), no external dependencies
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
- `save` - copies token from `Claude Code-credentials` to a named entry
- `login` - launches `claude` interactively, then saves the resulting token
- `use` - reads named token, sets `ANTHROPIC_AUTH_TOKEN`, execs `claude`
- `delete` - removes named keychain entry
- `show` - displays token preview (first/last 6 chars + length)
- `current` - reads `~/.claude/.active-context`

## Conventions

- No external Go dependencies (stdlib only)
- macOS-only (Keychain dependency)
- Exit codes: 0 = success, 1 = error
- Error messages go to stderr, output to stdout
- Binary name: `cx`
- Context names validated against `^[a-zA-Z0-9][a-zA-Z0-9._-]*$`

## Security

- gosec scans run in CI and locally via `make gosec`
- `#nosec` annotations are used for intentional patterns (exec.Command with hardcoded binary, syscall.Exec with LookPath result, fixed file paths)
- All errors must be handled — unhandled errors will fail gosec

## CI/CD

- `.github/workflows/ci.yml` - runs on push/PR to main: go vet, staticcheck, gofmt, gosec, build (amd64 + arm64)
- `.github/workflows/release.yml` - runs on `v*` tags: quality checks, builds stripped binaries, generates checksums and changelog, creates GitHub Release
- Build targets: `darwin/amd64` and `darwin/arm64` only
- Release binaries use `-ldflags="-s -w"` for size reduction

## Makefile targets

- `make lint` - run gofmt + go vet + staticcheck + gosec
- `make build` - build cx binary
- `make clean` - remove binaries
- `make` - lint + build (default)

## Testing changes

After modifying `cmd/cx/main.go`, verify:

```sh
make
./cx list && ./cx current
```

Do not run `cx save` or `cx delete` with real context names during testing. Use a disposable name like `test-tmp` and clean up after.

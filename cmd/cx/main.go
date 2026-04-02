package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
)

//go:embed completions/*
var completionFiles embed.FS

var (
	version   = "dev"
	validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
)

const (
	keychainSvc = "Claude Code-credentials"
)

var (
	account    string
	activeFile string
	verbose    bool
	debug      bool
)

func logVerbose(format string, a ...any) {
	if verbose || debug {
		fmt.Fprintf(os.Stderr, "[verbose] "+format+"\n", a...)
	}
}

func logDebug(format string, a ...any) {
	if debug {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", a...)
	}
}

func init() {
	u, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine current user: %v\n", err)
		os.Exit(1)
	}
	account = u.Username
	activeFile = filepath.Join(u.HomeDir, ".claude", ".active-context")
}

func svcName(name string) string {
	return fmt.Sprintf("%s (%s)", keychainSvc, name)
}

func keychainGet(service string) (string, error) {
	logDebug("keychain get: service=%q account=%q", service, account)
	out, err := exec.Command("security", "find-generic-password", // #nosec G204 -- args are validated or constant
		"-s", service,
		"-a", account,
		"-w").Output()
	if err != nil {
		logDebug("keychain get failed: %v", err)
		return "", err
	}
	token := strings.TrimSpace(string(out))
	logDebug("keychain get: got %d chars", len(token))
	return token, nil
}

func keychainSet(service, password string) error {
	logDebug("keychain set: service=%q account=%q token=%d chars", service, account, len(password))
	// Try update first
	err := exec.Command("security", "add-generic-password", // #nosec G204 -- args are validated or constant
		"-s", service,
		"-a", account,
		"-w", password,
		"-U").Run()
	if err != nil {
		// Fallback to add
		return exec.Command("security", "add-generic-password", // #nosec G204 -- args are validated or constant
			"-s", service,
			"-a", account,
			"-w", password).Run()
	}
	return nil
}

func keychainDelete(service string) error {
	logDebug("keychain delete: service=%q account=%q", service, account)
	return exec.Command("security", "delete-generic-password", // #nosec G204 -- args are validated or constant
		"-s", service,
		"-a", account).Run()
}

func listContexts() ([]string, error) {
	out, err := exec.Command("security", "dump-keychain").Output()
	if err != nil {
		return nil, err
	}

	prefix := keychainSvc + " ("
	seen := make(map[string]bool)
	var contexts []string

	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, `"svce"`) {
			continue
		}
		idx := strings.Index(line, prefix)
		if idx < 0 {
			continue
		}
		rest := line[idx+len(prefix):]
		end := strings.Index(rest, ")")
		if end < 0 {
			continue
		}
		name := rest[:end]
		if name != "" && !seen[name] {
			seen[name] = true
			contexts = append(contexts, name)
		}
	}
	sort.Strings(contexts)
	return contexts, nil
}

func getActiveContext() string {
	logDebug("reading active context from %s", activeFile)
	data, err := os.ReadFile(activeFile) // #nosec G304 -- path is fixed at init, not user-controlled
	if err != nil {
		logDebug("no active context file: %v", err)
		return ""
	}
	ctx := strings.TrimSpace(string(data))
	logDebug("active context: %q", ctx)
	return ctx
}

func setActiveContext(name string) error {
	logDebug("setting active context to %q in %s", name, activeFile)
	dir := filepath.Dir(activeFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(activeFile, []byte(name+"\n"), 0600) // #nosec G703 -- name is validated by caller, path is fixed
}

func cmdList() int {
	active := getActiveContext()
	contexts, err := listContexts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if len(contexts) == 0 {
		fmt.Println("No saved contexts. Use 'cx save <name>' or 'cx login <name>' to create one.")
		return 0
	}

	fmt.Println("Saved contexts:")
	for _, name := range contexts {
		if name == active {
			fmt.Printf("  * %s (active)\n", name)
		} else {
			fmt.Printf("    %s\n", name)
		}
	}
	return 0
}

func cmdSave(name string) int {
	logVerbose("saving current token as %q", name)
	token, err := keychainGet(keychainSvc)
	if err != nil || token == "" {
		fmt.Fprintln(os.Stderr, "Error: no token in keychain (not logged in?)")
		return 1
	}

	if err := keychainSet(svcName(name), token); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save token: %v\n", err)
		return 1
	}

	fmt.Printf("Saved current token as '%s'\n", name)
	return 0
}

func cmdLogin(name string) int {
	fmt.Printf("Logging in for context '%s'...\n", name)
	fmt.Println("Run /login in Claude, then exit.")
	fmt.Println()

	cmd := exec.Command("claude")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: claude exited with error: %v\n", err)
	}

	token, err := keychainGet(keychainSvc)
	if err != nil || token == "" {
		fmt.Fprintln(os.Stderr, "Error: no token found after login")
		return 1
	}

	if err := keychainSet(svcName(name), token); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save token: %v\n", err)
		return 1
	}

	if err := setActiveContext(name); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set active context: %v\n", err)
	}
	fmt.Printf("Context '%s' saved and active.\n", name)
	return 0
}

func cmdUse(name string, extra []string) int {
	logVerbose("switching to context %q", name)
	token, err := keychainGet(svcName(name))
	if err != nil || token == "" {
		fmt.Fprintf(os.Stderr, "Error: context '%s' not found\n", name)
		cmdList()
		return 1
	}

	if err := setActiveContext(name); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set active context: %v\n", err)
	}
	fmt.Printf("[%s]\n", name)

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: claude not found in PATH")
		return 1
	}
	logDebug("claude path: %s", claudePath)

	env := os.Environ()
	env = append(env, "ANTHROPIC_AUTH_TOKEN="+token)
	args := append([]string{"claude"}, extra...)
	logDebug("exec args: %v", args)

	return handleExec(claudePath, args, env)
}

func handleExec(path string, args, env []string) int {
	err := syscall.Exec(path, args, env) // #nosec G204 G702 -- path is from exec.LookPath("claude")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func cmdDelete(name string) int {
	logVerbose("deleting context %q", name)
	err := keychainDelete(svcName(name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: context '%s' not found\n", name)
		return 1
	}

	fmt.Printf("Deleted '%s'\n", name)

	if getActiveContext() == name {
		_ = os.Remove(activeFile)
	}
	return 0
}

func cmdShow(name string) int {
	token, err := keychainGet(svcName(name))
	if err != nil || token == "" {
		fmt.Fprintf(os.Stderr, "Error: context '%s' not found\n", name)
		return 1
	}

	l := len(token)
	prefix := token[:6]
	suffix := token[l-6:]
	fmt.Printf("%s: %s...%s (%d chars)\n", name, prefix, suffix, l)
	return 0
}

func cmdCurrent() int {
	active := getActiveContext()
	if active == "" {
		fmt.Println("default (keychain)")
	} else {
		fmt.Println(active)
	}
	return 0
}

func cmdCompletion(shell string) int {
	var file string
	switch shell {
	case "bash":
		file = "completions/cx.bash"
	case "zsh":
		file = "completions/cx.zsh"
	case "fish":
		file = "completions/cx.fish"
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported shell %q (use bash, zsh, or fish)\n", shell)
		return 1
	}
	data, err := completionFiles.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(data))
	return 0
}

func usage() {
	fmt.Print(`cx - Claude Code context switcher

Usage: cx [-v|--debug] <command> [args]

Commands:
  list                List all saved contexts
  save <name>         Save current keychain token as a named context
  login <name>        Login with a new account and save as context
  use <name>          Launch Claude Code with a saved context
  delete <name>       Delete a saved context
  show <name>         Show token preview (first/last chars)
  current             Show the currently active context
  completion <shell>  Print shell completions (bash, zsh, fish)
  version             Print version

Flags:
  -v                  Verbose output
  --debug             Debug output (includes verbose)

Examples:
  cx login green-code
  cx login personal
  cx use green-code
  cx -v list
  cx completion zsh > ~/.zfunc/_cx
`)
}

func requireName(args []string, cmd string) (string, bool) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: cx %s <name>\n", cmd)
		return "", false
	}
	name := args[0]
	if !validName.MatchString(name) {
		fmt.Fprintln(os.Stderr, "Error: name must be alphanumeric (hyphens, dots, underscores allowed)")
		return "", false
	}
	return name, true
}

func main() {
	args := os.Args[1:]

	// Parse flags
	for len(args) > 0 {
		switch args[0] {
		case "-v", "--verbose":
			verbose = true
			args = args[1:]
		case "--debug":
			debug = true
			args = args[1:]
		default:
			goto done
		}
	}
done:

	if debug {
		logDebug("account=%q activeFile=%q", account, activeFile)
	}

	if len(args) == 0 {
		usage()
		return
	}

	cmd := args[0]
	rest := args[1:]

	var code int
	switch cmd {
	case "list":
		code = cmdList()
	case "save":
		if name, ok := requireName(rest, cmd); ok {
			code = cmdSave(name)
		} else {
			code = 1
		}
	case "login":
		if name, ok := requireName(rest, cmd); ok {
			code = cmdLogin(name)
		} else {
			code = 1
		}
	case "use":
		if name, ok := requireName(rest, cmd); ok {
			code = cmdUse(name, rest[1:])
		} else {
			code = 1
		}
	case "delete":
		if name, ok := requireName(rest, cmd); ok {
			code = cmdDelete(name)
		} else {
			code = 1
		}
	case "show":
		if name, ok := requireName(rest, cmd); ok {
			code = cmdShow(name)
		} else {
			code = 1
		}
	case "current":
		code = cmdCurrent()
	case "version", "--version":
		fmt.Printf("cx %s\n", version)
	case "completion":
		if name, ok := requireName(rest, cmd); ok {
			code = cmdCompletion(name)
		} else {
			fmt.Fprintln(os.Stderr, "Usage: cx completion <bash|zsh|fish>")
			code = 1
		}
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown: %s\n", cmd)
		usage()
		code = 1
	}
	os.Exit(code)
}

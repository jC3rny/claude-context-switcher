package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSvcName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"work", "Claude Code-credentials (work)"},
		{"personal", "Claude Code-credentials (personal)"},
		{"a.b-c_d", "Claude Code-credentials (a.b-c_d)"},
	}
	for _, tt := range tests {
		if got := svcName(tt.name); got != tt.want {
			t.Errorf("svcName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestValidName(t *testing.T) {
	valid := []string{"foo", "foo-bar", "a.b", "a_b", "123", "A", "work1"}
	for _, name := range valid {
		if !validName.MatchString(name) {
			t.Errorf("validName should match %q", name)
		}
	}

	invalid := []string{"", "-start", ".start", "_start", "has space", "a/b", "a@b", "a(b)"}
	for _, name := range invalid {
		if validName.MatchString(name) {
			t.Errorf("validName should not match %q", name)
		}
	}
}

func TestGetSetActiveContext(t *testing.T) {
	orig := activeFile
	defer func() { activeFile = orig }()

	t.Run("no file", func(t *testing.T) {
		activeFile = filepath.Join(t.TempDir(), "missing")
		if got := getActiveContext(); got != "" {
			t.Errorf("getActiveContext() = %q, want empty", got)
		}
	})

	t.Run("round trip", func(t *testing.T) {
		activeFile = filepath.Join(t.TempDir(), "active")
		if err := setActiveContext("work"); err != nil {
			t.Fatalf("setActiveContext: %v", err)
		}
		if got := getActiveContext(); got != "work" {
			t.Errorf("getActiveContext() = %q, want %q", got, "work")
		}
	})

	t.Run("creates parent dirs", func(t *testing.T) {
		activeFile = filepath.Join(t.TempDir(), "sub", "dir", "active")
		if err := setActiveContext("test"); err != nil {
			t.Fatalf("setActiveContext: %v", err)
		}
		if got := getActiveContext(); got != "test" {
			t.Errorf("getActiveContext() = %q, want %q", got, "test")
		}
	})

	t.Run("overwrite", func(t *testing.T) {
		activeFile = filepath.Join(t.TempDir(), "active")
		_ = setActiveContext("first")
		if err := setActiveContext("second"); err != nil {
			t.Fatalf("setActiveContext: %v", err)
		}
		if got := getActiveContext(); got != "second" {
			t.Errorf("getActiveContext() = %q, want %q", got, "second")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "active")
		if err := os.WriteFile(f, []byte(""), 0600); err != nil {
			t.Fatal(err)
		}
		activeFile = f
		if got := getActiveContext(); got != "" {
			t.Errorf("getActiveContext() = %q, want empty", got)
		}
	})
}

func TestRequireName(t *testing.T) {
	tests := []struct {
		args    []string
		wantOK  bool
		wantVal string
	}{
		{[]string{}, false, ""},
		{[]string{"work"}, true, "work"},
		{[]string{"a-b.c_d"}, true, "a-b.c_d"},
		{[]string{"-bad"}, false, ""},
		{[]string{"has space"}, false, ""},
	}
	for _, tt := range tests {
		name, ok := requireName(tt.args, "test")
		if ok != tt.wantOK {
			t.Errorf("requireName(%v) ok = %v, want %v", tt.args, ok, tt.wantOK)
		}
		if ok && name != tt.wantVal {
			t.Errorf("requireName(%v) name = %q, want %q", tt.args, name, tt.wantVal)
		}
	}
}

func TestCmdCompletion(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		if code := cmdCompletion(shell); code != 0 {
			t.Errorf("cmdCompletion(%q) = %d, want 0", shell, code)
		}
	}
	if code := cmdCompletion("powershell"); code != 1 {
		t.Error("cmdCompletion(\"powershell\") should return 1")
	}
}

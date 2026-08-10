package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveConfigPath(t *testing.T) {
	t.Run("uses explicit path", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			if got := ResolveConfigPath("custom.yaml"); got != "custom.yaml" {
				t.Fatalf("ResolveConfigPath() = %q, want %q", got, "custom.yaml")
			}
		})
	})

	t.Run("prefers servers.yaml", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			mustWriteFile(t, "servers.yaml", "servers: []\n")
			mustWriteFile(t, "server.yml", "servers: []\n")

			if got := ResolveConfigPath(""); got != "servers.yaml" {
				t.Fatalf("ResolveConfigPath() = %q, want %q", got, "servers.yaml")
			}
		})
	})

	t.Run("falls back to legacy server.yml", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			mustWriteFile(t, "server.yml", "servers: []\n")

			if got := ResolveConfigPath(""); got != "server.yml" {
				t.Fatalf("ResolveConfigPath() = %q, want %q", got, "server.yml")
			}
		})
	})

	t.Run("defaults to servers.yaml when nothing exists", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			if got := ResolveConfigPath(""); got != "servers.yaml" {
				t.Fatalf("ResolveConfigPath() = %q, want %q", got, "servers.yaml")
			}
		})
	})
}

func TestValidateServers(t *testing.T) {
	t.Run("accepts valid config", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			cfg := Config{
				Servers: []Server{
					{
						UUID:  "server-1",
						Name:  "Server 1",
						Check: Action{Type: "ping", Params: []string{"127.0.0.1"}},
						Start: Action{Type: "wol", Params: []string{"00:11:22:33:44:55"}},
						Stop:  Action{Type: "ssh", Params: []string{"127.0.0.1", "shutdown now"}},
					},
				},
			}

			if err := ValidateServers(cfg); err != nil {
				t.Fatalf("ValidateServers() unexpected error: %v", err)
			}
		})
	})

	t.Run("rejects malformed params and missing fields", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			cfg := Config{
				Servers: []Server{
					{
						Check: Action{Type: "ping"},
						Start: Action{Type: "ipmi", Params: []string{"host", "user"}},
						Stop:  Action{Type: "ssh", Params: []string{"host"}},
					},
				},
			}

			err := ValidateServers(cfg)
			if err == nil {
				t.Fatal("ValidateServers() error = nil, want aggregated validation error")
			}

			msg := err.Error()
			for _, want := range []string{
				"uuid is required",
				"name is required",
				`type "ping" requires 1 params, got 0`,
				`type "ipmi" requires 3 params, got 2`,
				`type "ssh" requires 2 params, got 1`,
			} {
				if !strings.Contains(msg, want) {
					t.Fatalf("ValidateServers() error %q missing %q", msg, want)
				}
			}
		})
	})

	t.Run("rejects unknown non lua types", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			cfg := Config{
				Servers: []Server{
					{
						UUID:  "server-1",
						Name:  "Server 1",
						Check: Action{Type: "customcheck", Params: []string{"x"}},
						Start: Action{Type: "wol", Params: []string{"00:11:22:33:44:55"}},
						Stop:  Action{Type: "ssh", Params: []string{"127.0.0.1", "shutdown now"}},
					},
				},
			}

			err := ValidateServers(cfg)
			if err == nil || !strings.Contains(err.Error(), `unsupported type "customcheck"`) {
				t.Fatalf("ValidateServers() error = %v, want unsupported type", err)
			}
		})
	})

	t.Run("accepts custom lua types when script exists", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			mustWriteFile(t, filepath.Join("scripts", "custom", "customcheck.lua"), "print('success')\n")
			mustWriteFile(t, filepath.Join("scripts", "custom", "customstart.lua"), "print('success')\n")
			mustWriteFile(t, filepath.Join("scripts", "custom", "customstop.lua"), "print('success')\n")

			cfg := Config{
				Servers: []Server{
					{
						UUID:  "server-1",
						Name:  "Server 1",
						Check: Action{Type: "customcheck", Params: []string{"x"}},
						Start: Action{Type: "customstart"},
						Stop:  Action{Type: "customstop"},
					},
				},
			}

			if err := ValidateServers(cfg); err != nil {
				t.Fatalf("ValidateServers() unexpected error: %v", err)
			}
		})
	})
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q) error = %v", dir, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore cwd error = %v", err)
		}
	})

	fn()
}

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

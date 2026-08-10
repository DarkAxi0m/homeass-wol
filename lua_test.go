package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestRunLuaScript(t *testing.T) {
	if _, err := exec.LookPath("lua"); err != nil {
		t.Skip("lua executable not available")
	}

	t.Run("missing script", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			output, found, err := RunLuaScript("missing", nil)
			if found {
				t.Fatal("RunLuaScript() found = true, want false")
			}
			if err != nil {
				t.Fatalf("RunLuaScript() err = %v, want nil", err)
			}
			if output != "" {
				t.Fatalf("RunLuaScript() output = %q, want empty", output)
			}
		})
	})

	t.Run("script execution failure", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			mustWriteFile(t, "scripts/custom/fail.lua", "error('boom')\n")

			output, found, err := RunLuaScript("fail", nil)
			if !found {
				t.Fatal("RunLuaScript() found = false, want true")
			}
			if err == nil {
				t.Fatal("RunLuaScript() err = nil, want error")
			}
			if !strings.Contains(err.Error(), "lua script scripts/custom/fail.lua failed") {
				t.Fatalf("RunLuaScript() err = %v, want script path", err)
			}
			if output == "" {
				t.Fatal("RunLuaScript() output empty, want stderr/stdout content")
			}
		})
	})

	t.Run("script success", func(t *testing.T) {
		withWorkingDir(t, t.TempDir(), func() {
			mustWriteFile(t, "scripts/custom/pass.lua", "print('success')\n")

			output, found, err := RunLuaScript("pass", nil)
			if !found {
				t.Fatal("RunLuaScript() found = false, want true")
			}
			if err != nil {
				t.Fatalf("RunLuaScript() err = %v, want nil", err)
			}
			if output != "success" {
				t.Fatalf("RunLuaScript() output = %q, want %q", output, "success")
			}
		})
	})
}

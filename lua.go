package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func RunLuaScript(name string, params []string) (string, bool) {
	paths := []string{
		filepath.Join("scripts", "custom", name+".lua"),
		filepath.Join("scripts", "builtin", name+".lua"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			cmdArgs := append([]string{path}, params...)
			out, _ := exec.Command("lua", cmdArgs...).CombinedOutput()
			return string(out), true
		}
	}

	return "", false
}

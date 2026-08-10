package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func luaScriptPaths(name string) []string {
	return []string{
		filepath.Join("scripts", "custom", name+".lua"),
		filepath.Join("scripts", "builtin", name+".lua"),
	}
}

func FindLuaScript(name string) (string, bool) {
	for _, path := range luaScriptPaths(name) {
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func RunLuaScript(name string, params []string) (string, bool, error) {
	path, found := FindLuaScript(name)
	if !found {
		return "", false, nil
	}

	cmdArgs := append([]string{path}, params...)
	out, err := exec.Command("lua", cmdArgs...).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return output, true, fmt.Errorf("lua script %s failed: %w: %s", path, err, output)
	}

	return output, true, nil
}

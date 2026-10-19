package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// loadDotEnv sets KEY=VALUE pairs from path without overriding variables already
// in the environment. A missing file is not an error.
func loadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			// Report only the line number so a secret value never reaches the output
			return fmt.Errorf("%s:%d: expected KEY=VALUE", path, i+1)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, unquote(strings.TrimSpace(value))); err != nil {
			return fmt.Errorf("set %s from %s: %w", key, path, err)
		}
	}
	return nil
}

func unquote(value string) string {
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		return value[1 : len(value)-1]
	}
	return value
}

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnv accepts simple KEY=value files. Existing process variables win.
// No shell expansion/execution, and no values are printed in errors.
func LoadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && path == ".env" {
			return nil
		}
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		s := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		key, value, ok := strings.Cut(s, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return fmt.Errorf("invalid environment entry at line %d", line)
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set environment entry at line %d", line)
			}
		}
	}
	return scanner.Err()
}

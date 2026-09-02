package settings

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// values holds the current settings as key -> value pairs.
var values = map[string]string{}

// filePath is where settings are persisted.
var filePath string

// Init loads settings from ~/.taco/settings, creating the directory if needed.
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	filePath = filepath.Join(home, ".taco", "settings")

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return nil // no settings yet -> defaults
	}
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return scanner.Err()
}

// Get returns the value for key, or fallback if unset.
func Get(key, fallback string) string {
	if v, ok := values[key]; ok {
		return v
	}
	return fallback
}

// Set stores a value for key and persists all settings.
func Set(key, value string) error {
	values[key] = value
	return save()
}

// save writes all settings to disk (key=value per line, sorted).
func save() error {
	var b strings.Builder
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, values[k])
	}

	tmp := filePath + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, filePath)
}

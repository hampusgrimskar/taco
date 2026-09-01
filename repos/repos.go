package repos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Repos struct {
	repoMap map[string]string
	path    string
}

// Singleton pointer
var Instance *Repos

func Init() error {
	repos, err := loadRepos()

	if err != nil {
		return err
	}

	Instance = repos
	return nil
}

func loadRepos() (*Repos, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".taco", "repositories")

	r := &Repos{repoMap: make(map[string]string), path: path}

	if err := os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return r, nil // no file yet -> start empty
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		key, value, found := strings.Cut(line, "#")
		if !found {
			value = key
		}
		r.repoMap[key] = value
	}
	return r, scanner.Err()
}

func (repos *Repos) sync() error {
	var builder strings.Builder

	keys := make([]string, 0, len(repos.repoMap))
	for k := range repos.repoMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := repos.repoMap[key]

		if value == key {
			fmt.Fprintf(&builder, "%s\n", key) // no '#' when value == key
		} else {
			fmt.Fprintf(&builder, "%s#%s\n", key, value)
		}
	}

	tmp := repos.path + ".tmp"
	if err := os.WriteFile(tmp, []byte(builder.String()), 0644); err != nil {
		return err
	}

	return os.Rename(tmp, repos.path) // atomic swap
}

func (repos *Repos) Get() map[string]string {
	return repos.repoMap
}

func (repos *Repos) GetValue(key string) (string, bool) {
	value, ok := repos.repoMap[key]
	return value, ok
}

func (repos *Repos) SetValue(key string, value string) error {
	repos.repoMap[key] = value
	return repos.sync()
}

func (repos *Repos) DeleteKey(key string) error {
	delete(repos.repoMap, key)
	return repos.sync()
}

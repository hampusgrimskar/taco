package repos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hampusgrimskar/taco/session"
)

// Repo is a registered repository. Its Session is nil until a tmux
// session is started for it.
type Repo struct {
	Alias   string
	Path    string
	Session *session.Session
}

// Instance is the global list of registered repos.
var Instance []*Repo

// filePath is where the repo list is persisted.
var filePath string

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	filePath = filepath.Join(home, ".taco", "repositories")

	repos, err := loadRepos()
	if err != nil {
		return err
	}

	Instance = repos
	return nil
}

func loadRepos() ([]*Repo, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, err
	}

	repos := make([]*Repo, 0)

	f, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return repos, nil // no file yet -> start empty
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
		// File format is "path#alias": path before '#', alias after.
		path, alias, found := strings.Cut(line, "#")
		if !found {
			// No alias given: use the path as its own alias.
			alias = path
		}
		// Session starts nil; it is set when a tmux session is created.
		repos = append(repos, &Repo{Alias: alias, Path: path})
	}
	return repos, scanner.Err()
}

func sync() error {
	var builder strings.Builder

	// Write in alias order for stable output.
	sorted := make([]*Repo, len(Instance))
	copy(sorted, Instance)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Alias < sorted[j].Alias
	})

	for _, repo := range sorted {
		if repo.Alias == repo.Path {
			fmt.Fprintf(&builder, "%s\n", repo.Path) // no '#' when alias == path
		} else {
			fmt.Fprintf(&builder, "%s#%s\n", repo.Path, repo.Alias) // file format is path#alias
		}
	}

	tmp := filePath + ".tmp"
	if err := os.WriteFile(tmp, []byte(builder.String()), 0644); err != nil {
		return err
	}

	return os.Rename(tmp, filePath) // atomic swap
}

// Aliases returns all repo aliases as a sorted slice, suitable for
// feeding directly into a UI list.
func Aliases() []string {
	aliases := make([]string, 0, len(Instance))
	for _, repo := range Instance {
		aliases = append(aliases, repo.Alias)
	}
	sort.Strings(aliases)
	return aliases
}

// Find returns the repo with the given alias, or nil if none exists.
func Find(alias string) *Repo {
	for _, repo := range Instance {
		if repo.Alias == alias {
			return repo
		}
	}
	return nil
}

// WithSessions returns all repos that currently have a live session.
func WithSessions() []*Repo {
	active := make([]*Repo, 0)
	for _, repo := range Instance {
		if repo.Session != nil {
			active = append(active, repo)
		}
	}
	return active
}

// Add registers a new repo (or updates an existing alias) and persists the change.
func Add(alias string, path string) error {
	if r := Find(alias); r != nil {
		r.Path = path
	} else {
		Instance = append(Instance, &Repo{Alias: alias, Path: path})
	}
	return sync()
}

// Delete removes the repo with the given alias and persists the change.
func Delete(alias string) error {
	for i, repo := range Instance {
		if repo.Alias == alias {
			Instance = append(Instance[:i], Instance[i+1:]...)
			return sync()
		}
	}
	return nil
}

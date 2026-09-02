package repos

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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

// All returns the live repo list.
func All() []*Repo {
	return Instance
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

// Rename changes a repo's alias from oldAlias to newAlias and persists the
// change. It returns an error if newAlias is empty, if no repo has oldAlias,
// or if another repo already uses newAlias.
func Rename(oldAlias, newAlias string) error {
	newAlias = strings.TrimSpace(newAlias)
	if newAlias == "" {
		return fmt.Errorf("alias cannot be empty")
	}
	if newAlias == oldAlias {
		return nil // no-op
	}

	target := Find(oldAlias)
	if target == nil {
		return fmt.Errorf("no repo with alias %q", oldAlias)
	}
	if Find(newAlias) != nil {
		return fmt.Errorf("alias %q already in use", newAlias)
	}

	target.Alias = newAlias
	return sync()
}

// HasPath reports whether a repo with the given filesystem path is already
// registered.
func HasPath(path string) bool {
	for _, repo := range Instance {
		if repo.Path == path {
			return true
		}
	}
	return false
}

// skipDirs are directory names never descended into during a scan.
var skipDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	"vendor":       true,
}

// isGitRepo reports whether dir contains a .git entry (file or directory).
func isGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && (info.IsDir() || info.Mode().IsRegular())
}

// ScanGitRepos walks root recursively and returns the paths of all git repos
// found underneath it, sorted. It stops descending once a repo root is found
// (so nested repos inside a repo are not reported), and skips hidden and
// heavy directories.
func ScanGitRepos(root string) ([]string, error) {
	var found []string

	var walk func(dir string) error
	walk = func(dir string) error {
		if isGitRepo(dir) {
			found = append(found, dir)
			return nil // do not descend into a repo
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			// Unreadable directory (permissions, etc.): skip it, don't fail
			// the whole scan.
			return nil
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if skipDirs[name] || strings.HasPrefix(name, ".") {
				continue
			}
			if err := walk(filepath.Join(dir, name)); err != nil {
				return err
			}
		}
		return nil
	}

	// The root itself may be a repo.
	if err := walk(root); err != nil {
		return nil, err
	}
	sort.Strings(found)
	return found, nil
}

// ListDirs returns the immediate subdirectories of dir (names only, sorted),
// skipping hidden directories. Used by the directory browser.
func ListDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

// AddRepo registers the repo at path, deriving its alias from the directory
// name. If that alias is already taken, a numeric suffix is appended to make
// it unique (e.g. "web", "web-2"). If the path is already registered it is a
// no-op. It persists the change and returns the alias used (empty if skipped).
func AddRepo(path string) (string, error) {
	if HasPath(path) {
		return "", nil // already registered
	}

	base := filepath.Base(path)
	alias := base
	for n := 2; Find(alias) != nil; n++ {
		alias = fmt.Sprintf("%s-%d", base, n)
	}

	Instance = append(Instance, &Repo{Alias: alias, Path: path})
	if err := sync(); err != nil {
		return "", err
	}
	return alias, nil
}

// GitBranch returns the current branch name of the repo at path by reading
// .git/HEAD directly (no subprocess). For a detached HEAD it returns a short
// commit SHA. Returns "" if it cannot be determined.
func GitBranch(path string) string {
	data, err := os.ReadFile(filepath.Join(path, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	head := strings.TrimSpace(string(data))
	if ref, ok := strings.CutPrefix(head, "ref: refs/heads/"); ok {
		return ref
	}
	// Detached HEAD: head is a raw SHA.
	if len(head) >= 7 {
		return head[:7]
	}
	return head
}

// GitRecentCommits returns up to n recent commit subject lines for the repo at
// path, newest first. Returns nil (no error) if git is unavailable or the
// command fails, so callers can render gracefully.
func GitRecentCommits(path string, n int) []string {
	cmd := exec.Command("git", "-C", path, "log", fmt.Sprintf("-n%d", n), "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}

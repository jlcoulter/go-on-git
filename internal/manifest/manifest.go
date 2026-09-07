package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	FileName = "gog.toml"
	Version  = 1
)

type Repo struct {
	Path string `toml:"path"`
	URL  string `toml:"url,omitempty"`
}

type Manifest struct {
	Version int    `toml:"version"`
	BaseURL string `toml:"base_url,omitempty"`
	Repos   []Repo `toml:"repo"`
}

// Find locates the manifest. If explicit is non-empty it is used directly.
// Otherwise it walks up from startDir looking for gog.toml. It returns the
// manifest path and the directory containing it (the workspace root).
func Find(startDir, explicit string) (manifestPath, rootDir string, err error) {
	if explicit != "" {
		abs, aerr := filepath.Abs(explicit)
		if aerr != nil {
			return "", "", aerr
		}
		if _, serr := os.Stat(abs); serr != nil {
			return "", "", fmt.Errorf("manifest not found: %s", abs)
		}
		return abs, filepath.Dir(abs), nil
	}

	dir, derr := filepath.Abs(startDir)
	if derr != nil {
		return "", "", derr
	}
	for {
		candidate := filepath.Join(dir, FileName)
		if _, serr := os.Stat(candidate); serr == nil {
			return candidate, dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", fmt.Errorf("no %s found; run 'gog init' or 'gog scan' from your repos root", FileName)
}

func Load(path string) (*Manifest, error) {
	var m Manifest
	if _, err := toml.DecodeFile(path, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m.Version == 0 {
		m.Version = Version
	}
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &m, nil
}

func (m *Manifest) validate() error {
	seen := map[string]bool{}
	for _, r := range m.Repos {
		if r.Path == "" {
			return errors.New("repo entry has empty path")
		}
		if filepath.IsAbs(r.Path) {
			return fmt.Errorf("repo path must be relative: %q", r.Path)
		}
		norm := filepath.ToSlash(filepath.Clean(r.Path))
		if norm == "." || strings.HasPrefix(norm, "../") {
			return fmt.Errorf("repo path escapes workspace: %q", r.Path)
		}
		if seen[norm] {
			return fmt.Errorf("duplicate repo path: %q", norm)
		}
		seen[norm] = true
	}
	return nil
}

// Save writes the manifest atomically (temp file + rename), preserving order.
func (m *Manifest) Save(path string) error {
	var buf bytes.Buffer
	buf.WriteString("# gog manifest\n")
	if err := toml.NewEncoder(&buf).Encode(m); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Resolve returns the clone URL for a repo: explicit URL if set, otherwise
// derived from BaseURL and the path's basename.
func (m *Manifest) Resolve(r Repo) (string, error) {
	if r.URL != "" {
		return r.URL, nil
	}
	if m.BaseURL == "" {
		return "", fmt.Errorf("no url for %q and base_url is empty", r.Path)
	}
	return strings.TrimRight(m.BaseURL, "/") + "/" + filepath.Base(r.Path), nil
}

// Upsert adds repos not already present (by normalized path), appending new
// ones sorted alphabetically. Returns the repos that were added.
func (m *Manifest) Upsert(rs []Repo) []Repo {
	existing := map[string]bool{}
	for _, r := range m.Repos {
		existing[filepath.ToSlash(filepath.Clean(r.Path))] = true
	}
	var added []Repo
	for _, r := range rs {
		norm := filepath.ToSlash(filepath.Clean(r.Path))
		if existing[norm] {
			continue
		}
		r.Path = norm
		m.Repos = append(m.Repos, r)
		existing[norm] = true
		added = append(added, r)
	}
	sort.SliceStable(m.Repos, func(i, j int) bool {
		return m.Repos[i].Path < m.Repos[j].Path
	})
	return added
}

// Remove deletes entries by normalized path. Returns removed repos and any
// requested paths that were not present.
func (m *Manifest) Remove(paths []string) (removed []Repo, unknown []string) {
	want := map[string]bool{}
	for _, p := range paths {
		want[filepath.ToSlash(filepath.Clean(p))] = true
	}
	var kept []Repo
	for _, r := range m.Repos {
		norm := filepath.ToSlash(filepath.Clean(r.Path))
		if want[norm] {
			removed = append(removed, r)
			continue
		}
		kept = append(kept, r)
	}
	for p := range want {
		found := false
		for _, r := range removed {
			if filepath.ToSlash(filepath.Clean(r.Path)) == p {
				found = true
				break
			}
		}
		if !found {
			unknown = append(unknown, p)
		}
	}
	sort.Strings(unknown)
	m.Repos = kept
	return removed, unknown
}

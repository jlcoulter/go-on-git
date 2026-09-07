package manifest

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Scan walks root and returns the relative paths of every directory that
// contains a .git entry (file or dir, covering worktrees and submodules).
// It never descends into .git, prunes below discovered repos, and skips the
// root itself even if it is a repo.
func Scan(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == ".git" {
			return filepath.SkipDir
		}
		if path == root {
			return nil
		}
		if _, serr := os.Stat(filepath.Join(path, ".git")); serr == nil {
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				return rerr
			}
			paths = append(paths, filepath.ToSlash(rel))
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gog/internal/manifest"
)

// loadManifest resolves and loads the manifest using the --config flag or
// walk-up discovery from the current directory.
func loadManifest() (*manifest.Manifest, string, error) {
	cfgFile, _ := rootCmd.PersistentFlags().GetString("config")
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	mp, root, err := manifest.Find(cwd, cfgFile)
	if err != nil {
		return nil, "", err
	}
	m, err := manifest.Load(mp)
	if err != nil {
		return nil, "", err
	}
	return m, root, nil
}

// mustBool unwraps a (bool, error) flag lookup, panicking on error (flags are
// statically defined, so this never fails in practice).
func mustBool(v bool, err error) bool {
	if err != nil {
		panic(err)
	}
	return v
}

// selectRepos returns the manifest repos to operate on: those named in args
// (matched by path) if any, otherwise all repos.
func selectRepos(m *manifest.Manifest, args []string) ([]manifest.Repo, error) {
	if len(args) == 0 {
		return m.Repos, nil
	}
	want := map[string]bool{}
	for _, a := range args {
		want[filepath.ToSlash(filepath.Clean(a))] = true
	}
	var out []manifest.Repo
	for _, r := range m.Repos {
		if want[filepath.ToSlash(filepath.Clean(r.Path))] {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no manifest repos match: %v", args)
	}
	return out, nil
}

package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"gog/internal/manifest"
)

// isURL reports whether s looks like a remote URL (scheme:// or scp-like git@).
func isURL(s string) bool {
	return strings.Contains(s, "://") || strings.HasPrefix(s, "git@")
}

// addCmd declares repos in the manifest (for later cloning).
var addCmd = &cobra.Command{
	Use:   "add <name-or-url>...",
	Short: "Add repos to the manifest",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, root, err := loadManifest()
		if err != nil {
			return err
		}
		var repos []manifest.Repo
		for _, a := range args {
			r := manifest.Repo{Path: filepath.ToSlash(filepath.Clean(a))}
			if isURL(a) {
				r.URL = a
				r.Path = filepath.ToSlash(filepath.Base(strings.TrimSuffix(a, "/")))
			}
			repos = append(repos, r)
		}
		added := m.Upsert(repos)
		if err := m.Save(filepath.Join(root, manifest.FileName)); err != nil {
			return err
		}
		for _, r := range added {
			fmt.Printf("added %s\n", r.Path)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"gog/internal/manifest"
)

// removeCmd deletes entries from the manifest (never touches disk).
var removeCmd = &cobra.Command{
	Use:   "remove <path>...",
	Short: "Remove repos from the manifest",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, root, err := loadManifest()
		if err != nil {
			return err
		}
		removed, unknown := m.Remove(args)
		if err := m.Save(filepath.Join(root, manifest.FileName)); err != nil {
			return err
		}
		for _, r := range removed {
			fmt.Printf("removed %s\n", r.Path)
		}
		for _, u := range unknown {
			fmt.Printf("not in manifest: %s\n", u)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

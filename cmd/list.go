package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"gog/internal/gitx"
	"gog/internal/ui"
)

type listRow struct {
	Path   string `json:"path"`
	URL    string `json:"url"`
	Exists bool   `json:"exists"`
}

// listCmd prints manifest entries with their resolved URL and on-disk status.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List repos in the manifest",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, root, err := loadManifest()
		if err != nil {
			return err
		}
		onlyMissing, _ := cmd.Flags().GetBool("missing")
		jsonOut, _ := rootCmd.PersistentFlags().GetBool("json")

		var rows []listRow
		for _, r := range m.Repos {
			url, _ := m.Resolve(r)
			exists := gitx.IsRepo(filepath.Join(root, r.Path))
			if onlyMissing && exists {
				continue
			}
			rows = append(rows, listRow{Path: r.Path, URL: url, Exists: exists})
		}

		if jsonOut {
			return ui.PrintJSON(map[string]any{"command": "list", "repos": rows})
		}
		t := ui.NewTable("PATH", "URL", "EXISTS")
		for _, r := range rows {
			exists := "yes"
			if !r.Exists {
				exists = "missing"
			}
			t.AddRow(r.Path, r.URL, exists)
		}
		t.Render(cmd.OutOrStdout())
		return nil
	},
}

func init() {
	listCmd.Flags().Bool("missing", false, "only list repos missing from disk")
	rootCmd.AddCommand(listCmd)
}

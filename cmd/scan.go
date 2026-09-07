package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"gog/internal/gitx"
	"gog/internal/manifest"
	"gog/internal/ui"
)

type scanRow struct {
	Path   string `json:"path"`
	Result string `json:"result"`
}

// scanCmd discovers repos under the workspace root and registers them.
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Discover repos under the workspace and register them in the manifest",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		jsonOut, _ := rootCmd.PersistentFlags().GetBool("json")

		cfgFile, _ := rootCmd.PersistentFlags().GetString("config")
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		mp, root, err := manifest.Find(cwd, cfgFile)
		created := false
		if err != nil {
			// No manifest: treat CWD as the workspace root and create one.
			root = cwd
			mp = filepath.Join(root, manifest.FileName)
			created = true
		}

		paths, err := manifest.Scan(root)
		if err != nil {
			return err
		}

		var rows []scanRow
		var repos []manifest.Repo
		for _, p := range paths {
			url, _ := gitx.OriginURL(filepath.Join(root, p))
			repos = append(repos, manifest.Repo{Path: p, URL: url})
			rows = append(rows, scanRow{Path: p, Result: "added"})
		}

		if !dryRun {
			m := &manifest.Manifest{Version: manifest.Version}
			if !created {
				if m, err = manifest.Load(mp); err != nil {
					return err
				}
			}
			added := m.Upsert(repos)
			addedSet := map[string]bool{}
			for _, r := range added {
				addedSet[r.Path] = true
			}
			for i := range rows {
				if !addedSet[rows[i].Path] {
					rows[i].Result = "exists"
				}
			}
			if err := m.Save(mp); err != nil {
				return err
			}
		}

		if jsonOut {
			return ui.PrintJSON(map[string]any{"command": "scan", "repos": rows})
		}
		t := ui.NewTable("PATH", "RESULT")
		for _, r := range rows {
			t.AddRow(r.Path, r.Result)
		}
		t.Render(cmd.OutOrStdout())
		return nil
	},
}

func init() {
	scanCmd.Flags().Bool("dry-run", false, "print what would change without writing")
	rootCmd.AddCommand(scanCmd)
}

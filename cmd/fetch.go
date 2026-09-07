package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"gog/internal/gitx"
	"gog/internal/manifest"
	"gog/internal/runner"
	"gog/internal/ui"
)

type fetchRow struct {
	Path   string `json:"path"`
	Result string `json:"result"`
	Detail string `json:"detail,omitempty"`
}

// fetchCmd fetches all remotes in each repo.
var fetchCmd = &cobra.Command{
	Use:   "fetch [path...]",
	Short: "Fetch all remotes in each repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, root, err := loadManifest()
		if err != nil {
			return err
		}
		repos, err := selectRepos(m, args)
		if err != nil {
			return err
		}
		parallel, _ := cmd.Flags().GetInt("parallel")
		quiet, _ := cmd.Flags().GetBool("quiet")
		jsonOut, _ := rootCmd.PersistentFlags().GetBool("json")
		color := ui.NewColor(mustBool(rootCmd.PersistentFlags().GetBool("no-color")))
		progress := ui.NewProgress(cmd.ErrOrStderr(), len(repos), "fetching")

		rows := runner.Map(repos, parallel, func(r manifest.Repo) fetchRow {
			dir := filepath.Join(root, r.Path)
			if !gitx.IsRepo(dir) {
				return fetchRow{Path: r.Path, Result: "skipped", Detail: "missing"}
			}
			detail, ferr := gitx.Fetch(dir)
			if ferr != nil {
				return fetchRow{Path: r.Path, Result: "error", Detail: detail}
			}
			return fetchRow{Path: r.Path, Result: "ok"}
		}, progress.Inc)
		progress.Done()

		failed := 0
		for _, r := range rows {
			if r.Result == "error" {
				failed++
			}
		}
		if jsonOut {
			return ui.PrintJSON(map[string]any{"command": "fetch", "repos": rows, "failed": failed})
		}
		t := ui.NewTable("PATH", "RESULT", "DETAIL")
		for _, r := range rows {
			if quiet && r.Result != "error" {
				continue
			}
			res := r.Result
			switch res {
			case "ok":
				res = color.Green(res)
			case "error":
				res = color.Red(res)
			case "skipped":
				res = color.Yellow(res)
			}
			t.AddRow(r.Path, res, r.Detail)
		}
		t.Render(cmd.OutOrStdout())
		if failed > 0 {
			return fmt.Errorf("%d of %d repos failed to fetch", failed, len(rows))
		}
		return nil
	},
}

func init() {
	fetchCmd.Flags().IntP("parallel", "p", 4, "number of parallel fetches")
	fetchCmd.Flags().BoolP("quiet", "q", false, "only print failures")
	rootCmd.AddCommand(fetchCmd)
}

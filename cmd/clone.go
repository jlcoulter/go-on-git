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

type cloneRow struct {
	Path   string `json:"path"`
	Result string `json:"result"`
	Detail string `json:"detail,omitempty"`
	URL    string `json:"url,omitempty"`
}

// cloneCmd clones manifest repos that are missing from disk.
var cloneCmd = &cobra.Command{
	Use:   "clone [path...]",
	Short: "Clone repos declared in the manifest that are missing from disk",
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
		progress := ui.NewProgress(cmd.ErrOrStderr(), len(repos), "cloning")

		type job struct {
			repo manifest.Repo
			row  cloneRow
		}
		jobs := make([]job, 0, len(repos))
		for _, r := range repos {
			dest := filepath.Join(root, r.Path)
			if gitx.IsRepo(dest) {
				jobs = append(jobs, job{r, cloneRow{Path: r.Path, Result: "exists", Detail: "skipped"}})
				continue
			}
			url, rerr := m.Resolve(r)
			if rerr != nil {
				jobs = append(jobs, job{r, cloneRow{Path: r.Path, Result: "error", Detail: rerr.Error()}})
				continue
			}
			jobs = append(jobs, job{r, cloneRow{Path: r.Path, Result: "cloning", URL: url}})
		}

		results := runner.Map(jobs, parallel, func(j job) cloneRow {
			if j.row.Result != "cloning" {
				return j.row
			}
			detail, cerr := gitx.Clone(j.row.URL, filepath.Join(root, j.repo.Path))
			if cerr != nil {
				return cloneRow{Path: j.repo.Path, Result: "error", Detail: detail}
			}
			return cloneRow{Path: j.repo.Path, Result: "ok"}
		}, progress.Inc)
		progress.Done()

		failed := 0
		for _, r := range results {
			if r.Result == "error" {
				failed++
			}
		}
		if jsonOut {
			return ui.PrintJSON(map[string]any{"command": "clone", "repos": results, "failed": failed})
		}
		t := ui.NewTable("PATH", "RESULT", "DETAIL")
		for _, r := range results {
			if quiet && r.Result != "error" {
				continue
			}
			res := r.Result
			switch res {
			case "ok":
				res = color.Green(res)
			case "error":
				res = color.Red(res)
			case "exists":
				res = color.Yellow(res)
			}
			t.AddRow(r.Path, res, r.Detail)
		}
		t.Render(cmd.OutOrStdout())
		if failed > 0 {
			return fmt.Errorf("%d of %d repos failed to clone", failed, len(results))
		}
		return nil
	},
}

func init() {
	cloneCmd.Flags().IntP("parallel", "p", 4, "number of parallel clones")
	cloneCmd.Flags().BoolP("quiet", "q", false, "only print failures")
	rootCmd.AddCommand(cloneCmd)
}

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

type pullRow struct {
	Path   string `json:"path"`
	Result string `json:"result"`
	Detail string `json:"detail,omitempty"`
}

// pullCmd fast-forwards each repo to its upstream.
var pullCmd = &cobra.Command{
	Use:   "pull [path...]",
	Short: "Pull (fast-forward) each repo to its upstream",
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
		rebase, _ := cmd.Flags().GetBool("rebase")
		jsonOut, _ := rootCmd.PersistentFlags().GetBool("json")
		color := ui.NewColor(mustBool(rootCmd.PersistentFlags().GetBool("no-color")))
		progress := ui.NewProgress(cmd.ErrOrStderr(), len(repos), "pulling")

		rows := runner.Map(repos, parallel, func(r manifest.Repo) pullRow {
			dir := filepath.Join(root, r.Path)
			if !gitx.IsRepo(dir) {
				return pullRow{Path: r.Path, Result: "skipped", Detail: "missing"}
			}
			st, serr := gitx.Status(dir)
			if serr != nil {
				return pullRow{Path: r.Path, Result: "error", Detail: serr.Error()}
			}
			if st.Detached || st.Unborn {
				return pullRow{Path: r.Path, Result: "skipped", Detail: "no branch"}
			}
			if st.Upstream == "" {
				return pullRow{Path: r.Path, Result: "skipped", Detail: "no upstream"}
			}
			detail, perr := gitx.Pull(dir, rebase)
			if perr != nil {
				return pullRow{Path: r.Path, Result: "error", Detail: detail}
			}
			return pullRow{Path: r.Path, Result: "ok", Detail: "up to date"}
		}, progress.Inc)
		progress.Done()

		failed := 0
		for _, r := range rows {
			if r.Result == "error" {
				failed++
			}
		}
		if jsonOut {
			return ui.PrintJSON(map[string]any{"command": "pull", "repos": rows, "failed": failed})
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
			return fmt.Errorf("%d of %d repos failed to pull", failed, len(rows))
		}
		return nil
	},
}

func init() {
	pullCmd.Flags().IntP("parallel", "p", 4, "number of parallel pulls")
	pullCmd.Flags().BoolP("quiet", "q", false, "only print failures")
	pullCmd.Flags().Bool("rebase", false, "pull with --rebase instead of --ff-only")
	rootCmd.AddCommand(pullCmd)
}

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

type statusRow struct {
	Path      string `json:"path"`
	Branch    string `json:"branch"`
	Upstream  string `json:"upstream,omitempty"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	Dirty     int    `json:"dirty"`
	Untracked int    `json:"untracked"`
	State     string `json:"state"`
	Error     string `json:"error,omitempty"`
}

// statusCmd reports the state of every on-disk repo.
var statusCmd = &cobra.Command{
	Use:   "status [path...]",
	Short: "Show branch and working-tree state for each repo",
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
		progress := ui.NewProgress(cmd.ErrOrStderr(), len(repos), "checking status")

		rows := runner.Map(repos, parallel, func(r manifest.Repo) statusRow {
			dir := filepath.Join(root, r.Path)
			if !gitx.IsRepo(dir) {
				return statusRow{Path: r.Path, State: "missing"}
			}
			st, serr := gitx.Status(dir)
			if serr != nil {
				return statusRow{Path: r.Path, State: "error", Error: serr.Error()}
			}
			return statusRow{
				Path: r.Path, Branch: st.Branch, Upstream: st.Upstream,
				Ahead: st.Ahead, Behind: st.Behind, Dirty: st.Dirty,
				Untracked: st.Untracked, State: st.State(),
			}
		}, progress.Inc)
		progress.Done()

		if jsonOut {
			return ui.PrintJSON(map[string]any{"command": "status", "repos": rows})
		}
		t := ui.NewTable("PATH", "BRANCH", "AHEAD", "BEHIND", "DIRTY", "UNTRACKED", "STATE")
		for _, r := range rows {
			if quiet && (r.State == "clean" || r.State == "missing") {
				continue
			}
			state := r.State
			switch state {
			case "clean":
				state = color.Green(state)
			case "missing", "error":
				state = color.Red(state)
			default:
				state = color.Yellow(state)
			}
			t.AddRow(r.Path, r.Branch, fmt.Sprint(r.Ahead), fmt.Sprint(r.Behind),
				fmt.Sprint(r.Dirty), fmt.Sprint(r.Untracked), state)
		}
		t.Render(cmd.OutOrStdout())
		return nil
	},
}

func init() {
	statusCmd.Flags().IntP("parallel", "p", 4, "number of parallel status checks")
	statusCmd.Flags().BoolP("quiet", "q", false, "only print non-clean repos")
	rootCmd.AddCommand(statusCmd)
}

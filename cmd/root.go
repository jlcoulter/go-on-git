package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"gog/internal/manifest"
)

var rootCmd = &cobra.Command{
	Use:   "gog",
	Short: "Recursive git repository manager",
	Long: `gog discovers, registers, clones, and operates on every git repository
under a workspace root. A workspace is a directory containing a gog.toml
manifest; all repo paths are relative to it.

Commands run batch git operations in parallel and report results as an
aligned table, a quiet failure list, or JSON.`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "path to gog.toml manifest (default: walk up from CWD)")
	rootCmd.PersistentFlags().Bool("json", false, "emit JSON output")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
}

// initCmd creates a fresh manifest in the current directory.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new gog.toml manifest in the current directory",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path := filepath.Join(cwd, manifest.FileName)
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		m := &manifest.Manifest{Version: manifest.Version}
		if err := m.Save(path); err != nil {
			return err
		}
		fmt.Printf("created %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

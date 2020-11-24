package cmd

import (
	"github.com/spf13/cobra"

	"github.com/yunionio/git-tools/pkg/changelog-gen/cmd/config"
	"github.com/yunionio/git-tools/pkg/changelog-gen/cmd/run"
)

var (
	rootCmd = &cobra.Command{
		Use:   "changelog-gen",
		Short: "Changelog generator for yunionio open source projects",
	}
)

func init() {
	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(run.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}

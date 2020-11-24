package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"yunion.io/x/jsonutils"

	"github.com/yunionio/git-tools/pkg/types"
)

var (
	Cmd = &cobra.Command{
		Use:   "config",
		Short: "Config related actions",
	}

	exampleCmd = &cobra.Command{
		Use:   "example",
		Short: "Show example config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showExampleConfig()
		},
	}
)

func init() {
	Cmd.AddCommand(exampleCmd)
}

func showExampleConfig() error {
	fmt.Printf(jsonutils.Marshal(types.ExampleGlocalChangeLogConfigv1).YAMLString())
	return nil
}

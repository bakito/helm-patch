package cmd

import (
	"errors"
	"io"

	"github.com/spf13/cobra"
)

var (
	settings *envSettings
)

// NewRootCmd create a new root command
func NewRootCmd(out io.Writer, args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "patch",
		Short:        "Patch helm 3 releases",
		Long:         "Patch helm 3 releases",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("no arguments accepted")
			}
			return nil
		},
	}

	flags := cmd.PersistentFlags()
	flags.Parse(args)
	settings = newEnvSettings()

	cmd.AddCommand(
		newAPICmd(out),
		newAdoptCmd(out),
	)

	return cmd
}

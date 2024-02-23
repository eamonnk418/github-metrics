package util

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func ParseLocalCmdFlags(cmd *cobra.Command, args []string) error {
	return cmd.Flags().ParseAll(args, func(flag *pflag.Flag, value string) error {
		return flag.Value.Set(value)
	})
}
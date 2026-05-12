package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd(ver string) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update ariavox to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Phase 5 stub: will call updater package (same pattern as chop)
			fmt.Printf("ariavox update: not yet implemented\ncurrent version: %s\n", ver)
			return nil
		},
	}
}

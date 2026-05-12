package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute(ver string) {
	rootCmd := newRootCmd(ver)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd(ver string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ariavox",
		Short:        "Accessible PTY wrapper for AI coding agents",
		Long:         "ariavox intercepts terminal I/O from AI coding agents and makes it accessible to screen readers and TTS engines.",
		SilenceUsage: true,
	}

	cmd.AddCommand(
		newRunCmd(),
		newConfigCmd(),
		newDoctorCmd(),
		newUpdateCmd(ver),
		newVersionCmd(ver),
	)

	return cmd
}

func newVersionCmd(ver string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ariavox %s\n", ver)
		},
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var srMode  bool
	var ttsMode bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "run -- <command> [args...]",
		Short: "Run an AI agent with accessible output",
		Example: `  ariavox run -- claude
  ariavox run --sr -- claude --dangerously-skip-permissions
  ariavox run --tts -- claude`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Phase 1 stub: will be wired to internal/pty
			fmt.Fprintf(os.Stderr, "ariavox: run not yet implemented (phase 1)\n")
			fmt.Fprintf(os.Stderr, "command: %v sr=%v tts=%v verbose=%v\n", args, srMode, ttsMode, verbose)
			return nil
		},
	}

	cmd.Flags().BoolVar(&srMode, "sr", false, "Enable screen reader mode (strip ANSI, suppress spinners)")
	cmd.Flags().BoolVar(&ttsMode, "tts", false, "Enable TTS announcements")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging to stderr")

	return cmd
}

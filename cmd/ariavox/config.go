package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage ariavox configuration",
	}

	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigSetCmd(),
		newConfigEditCmd(),
		newConfigPathCmd(),
	)

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Phase 5 stub: will be wired to internal/config
			fmt.Println("ariavox config show: not yet implemented (phase 5)")
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set <key> <value>",
		Short:   "Set a configuration value",
		Example: "  ariavox config set tts.rate 60",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("ariavox config set: not yet implemented (phase 5)\nkey=%s value=%s\n", args[0], args[1])
			return nil
		},
	}
}

func newConfigEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open configuration file in editor",
		Long: `Opens the configuration file in the editor defined by $VISUAL or $EDITOR.
Falls back to vim, nano, or notepad in that order.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := configFilePath()
			if err != nil {
				return err
			}

			// Ensure config file exists
			if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
				if err := ensureDefaultConfig(cfgPath); err != nil {
					return fmt.Errorf("could not create config file: %w", err)
				}
			}

			editor := resolveEditor()
			editorCmd := exec.Command(editor, cfgPath)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr
			return editorCmd.Run()
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the path to the configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := configFilePath()
			if err != nil {
				return err
			}
			fmt.Println(p)
			return nil
		},
	}
}

// resolveEditor returns the first available editor from env vars and fallbacks.
func resolveEditor() string {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if e := os.Getenv(env); e != "" {
			return e
		}
	}
	for _, e := range []string{"vim", "nano", "vi", "notepad"} {
		if _, err := exec.LookPath(e); err == nil {
			return e
		}
	}
	return "vi"
}

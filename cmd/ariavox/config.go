package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"

	"github.com/AgusRdz/ariavox/internal/config"
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
			cfgPath, err := configFilePath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			out, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}
			fmt.Printf("# config file: %s\n", cfgPath)
			fmt.Print(string(out))
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set <key> <value>",
		Short:   "Set a configuration value",
		Example: "  ariavox config set tts.rate 60\n  ariavox config set screen_reader.enabled true",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := configFilePath()
			if err != nil {
				return err
			}
			if err := ensureDefaultConfig(cfgPath); err != nil {
				return err
			}
			if err := config.Set(cfgPath, args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "set %s = %s\n", args[0], args[1])
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
			if err := ensureDefaultConfig(cfgPath); err != nil {
				return fmt.Errorf("could not create config file: %w", err)
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

const defaultConfigYAML = `# ariavox configuration
# Edit with: ariavox config edit
# Full reference: https://github.com/AgusRdz/ariavox

tts:
  enabled: false
  rate: 50        # 0-100
  volume: 80      # 0-100
  voice: ""       # empty = system default
  priority_filter:
    - task_end
    - error
    - tool_use

screen_reader:
  enabled: false        # set true or use --sr flag / ARIAVOX_SR=1
  strip_ansi: true
  suppress_spinners: true

renderer:
  separators: true
  semantic_prefixes: true
  high_contrast: false
  respect_no_color: true

agent:
  command: ["claude"]
  args: []
`

func ensureDefaultConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(defaultConfigYAML)
	return err
}

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/AgusRdz/ariavox/internal/config"
	"github.com/AgusRdz/ariavox/internal/processor"
	"github.com/AgusRdz/ariavox/internal/pty"
	"github.com/AgusRdz/ariavox/internal/renderer"
	"github.com/AgusRdz/ariavox/internal/srbus"
	"github.com/AgusRdz/ariavox/internal/tts"
	"github.com/AgusRdz/ariavox/pkg/ansi"
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
			cfgPath, err := configFilePath()
			if err != nil {
				return err
			}
			cfg, err := config.Load(cfgPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ariavox: config warning: %v\n", err)
				cfg = config.Default()
			}
			// CLI flags and env vars override config file
			if os.Getenv("ARIAVOX_SR") != "" {
				srMode = true
			} else if cmd.Flags().Changed("sr") {
				cfg.ScreenReader.Enabled = srMode
			} else {
				srMode = cfg.ScreenReader.Enabled
			}
			if cmd.Flags().Changed("tts") {
				cfg.TTS.Enabled = ttsMode
			} else {
				ttsMode = cfg.TTS.Enabled
			}
			return runAgent(args, cfg, srMode, ttsMode, verbose)
		},
	}

	cmd.Flags().BoolVar(&srMode, "sr", false, "Enable screen reader mode (strip ANSI, suppress spinners)")
	cmd.Flags().BoolVar(&ttsMode, "tts", false, "Enable TTS announcements")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging to stderr")

	return cmd
}

func runAgent(args []string, cfg config.Config, srMode, ttsMode, verbose bool) error {
	p := pty.New()

	child := exec.Command(args[0], args[1:]...)

	if err := p.Start(child); err != nil {
		return fmt.Errorf("ariavox: %w", err)
	}
	defer p.Close()

	if rows, cols, err := termSize(); err == nil {
		_ = p.Resize(rows, cols)
	}

	proc := processor.New()

	rCfg := renderer.Config{
		SRMode:           srMode,
		Separators:       cfg.Renderer.Separators,
		SemanticPrefixes: cfg.Renderer.SemanticPrefixes,
		HighContrast:     cfg.Renderer.HighContrast,
	}
	if cfg.Renderer.RespectNoColor && os.Getenv("NO_COLOR") != "" {
		rCfg.HighContrast = true
	}
	rend := renderer.New(rCfg, os.Stdout)

	var bridge *tts.Bridge
	if ttsMode {
		spk := tts.New()
		_ = spk.SetRate(cfg.TTS.Rate)
		_ = spk.SetVolume(cfg.TTS.Volume)
		spk.SetVoice(cfg.TTS.Voice)
		bridge = tts.NewBridge(spk, cfg.TTSKindFilter())
		defer bridge.Close()
	}

	var bus srbus.Bus
	if srMode {
		bus = srbus.New()
		defer bus.Close()
	} else {
		bus = srbus.NopBus{}
	}

	parser := &ansi.Parser{}

	go func() {
		_, _ = io.Copy(p, os.Stdin)
	}()

	sigCh := make(chan os.Signal, 8)
	notifySignals(sigCh)
	go handleSignals(sigCh, p)

	buf := make([]byte, 4096)
	for {
		n, err := p.Read(buf)
		if n > 0 {
			chunk := buf[:n]

			_, _ = os.Stdout.Write(chunk)

			if ttsMode || srMode {
				clean := parser.Process(chunk)
				events := proc.Process(clean)
				for _, ev := range events {
					if srMode {
						rend.Render(ev)
						if ev.Text != "" {
							if busErr := bus.Announce(ev.Text); busErr != nil && verbose {
								fmt.Fprintf(os.Stderr, "ariavox: srbus: %v\n", busErr)
							}
						}
					}
					if ttsMode && bridge != nil {
						if verbose {
							fmt.Fprintf(os.Stderr, "ariavox: event kind=%d priority=%d text=%q\n", ev.Kind, ev.Priority, ev.Text)
						}
						var logf func(string, ...any)
						if verbose {
							logf = func(f string, a ...any) {
								fmt.Fprintf(os.Stderr, "ariavox: "+f+"\n", a...)
							}
						}
						bridge.Send(ev, logf)
					}
				}
			}
		}
		if err != nil {
			if isReadDone(err) {
				break
			}
			if verbose {
				fmt.Fprintf(os.Stderr, "ariavox: read: %v\n", err)
			}
			break
		}
	}

	signal.Stop(sigCh)
	close(sigCh)

	waitErr := p.Wait()
	p.Close() // restore terminal state before os.Exit (defers don't run on os.Exit)
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	return waitErr
}

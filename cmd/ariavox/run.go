package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/AgusRdz/ariavox/internal/processor"
	"github.com/AgusRdz/ariavox/internal/pty"
	"github.com/AgusRdz/ariavox/internal/renderer"
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
			return runAgent(args, srMode, ttsMode, verbose)
		},
	}

	cmd.Flags().BoolVar(&srMode, "sr", false, "Enable screen reader mode (strip ANSI, suppress spinners)")
	cmd.Flags().BoolVar(&ttsMode, "tts", false, "Enable TTS announcements")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging to stderr")

	return cmd
}

func runAgent(args []string, srMode, ttsMode, verbose bool) error {
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

	rCfg := renderer.DefaultConfig()
	rCfg.SRMode = srMode
	rend := renderer.New(rCfg, os.Stdout)

	var speaker tts.Speaker
	if ttsMode {
		speaker = tts.New()
		defer speaker.Close()
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
					}
					if ttsMode && speaker != nil && ev.Text != "" {
						if speakErr := speaker.Speak(ev.Text, ev.Priority); speakErr != nil && verbose {
							fmt.Fprintf(os.Stderr, "ariavox: tts: %v\n", speakErr)
						}
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
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	return waitErr
}

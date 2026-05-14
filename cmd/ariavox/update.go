package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const repo = "AgusRdz/ariavox"

func newUpdateCmd(ver string) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update ariavox to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(ver)
		},
	}
}

func runUpdate(current string) error {
	fmt.Println("checking for updates...")

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("could not fetch latest version: %w", err)
	}

	if latest == current {
		fmt.Printf("already up to date (%s)\n", current)
		return nil
	}

	fmt.Printf("updating %s → %s\n", current, latest)

	binary := binaryName()
	baseURL := fmt.Sprintf("https://github.com/%s/releases/download/%s", repo, latest)
	binaryURL := baseURL + "/" + binary
	checksumURL := baseURL + "/checksums.txt"

	// Download binary to temp file
	tmp, err := os.CreateTemp("", "ariavox-update-*")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	fmt.Printf("downloading %s...\n", binary)
	if err := downloadFile(binaryURL, tmp); err != nil {
		tmp.Close()
		return fmt.Errorf("download failed: %w", err)
	}
	tmp.Close()

	// Verify checksum
	checksums, err := fetchChecksums(checksumURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not fetch checksums — skipping verification: %v\n", err)
	} else {
		expected, ok := checksums[binary]
		if !ok {
			fmt.Fprintf(os.Stderr, "warning: %s not found in checksums.txt — skipping verification\n", binary)
		} else {
			actual, err := sha256File(tmpPath)
			if err != nil {
				return fmt.Errorf("checksum: %w", err)
			}
			if actual != expected {
				return fmt.Errorf("checksum mismatch\n  expected: %s\n  actual:   %s", expected, actual)
			}
			fmt.Println("checksum verified")
		}
	}

	// Replace current executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}

	if err := replaceExecutable(tmpPath, exe); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	fmt.Printf("updated to %s\n", latest)
	return nil
}

func fetchLatestVersion() (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("empty tag_name in response")
	}
	return payload.TagName, nil
}

func binaryName() string {
	switch runtime.GOOS {
	case "darwin":
		return "ariavox-darwin-universal"
	case "windows":
		arch := "amd64"
		if runtime.GOARCH == "arm64" {
			arch = "arm64"
		}
		return "ariavox-windows-" + arch + ".exe"
	default: // linux
		arch := runtime.GOARCH
		return "ariavox-linux-" + arch
	}
}

func downloadFile(url string, dst *os.File) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	_, err = io.Copy(dst, resp.Body)
	return err
}

func fetchChecksums(url string) (map[string]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			name := strings.TrimPrefix(parts[1], "*")
			m[name] = parts[0]
		}
	}
	return m, nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func replaceExecutable(src, dst string) error {
	dir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dir, ".ariavox-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	in, err := os.Open(src)
	if err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	_, err = io.Copy(tmp, in)
	in.Close()
	tmp.Close()
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return replaceExe(tmpPath, dst)
}

// cleanOldExe removes any leftover .old backup from a previous update.
// Called at startup so it runs after the old process has exited.
func cleanOldExe() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	old := exe + ".old"
	_ = os.Remove(old)
}

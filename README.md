# ariavox

**Accessible PTY wrapper for AI coding agents.**

Modern AI agent CLIs like Claude Code stream output using spinners, ANSI color codes, in-place cursor updates, and Unicode decoration. These break screen readers (NVDA, JAWS, VoiceOver, Orca) — producing garbled speech, frozen output, or silence. ariavox fixes this at the PTY layer without requiring any changes to the agent.

```
ariavox run -- claude
```

That's it. ariavox wraps the agent process, intercepts its output, strips everything that confuses screen readers, and delivers clean text to your OS TTS engine and/or screen reader.

## Features

- **Screen reader mode** (`--sr`) — strips all ANSI sequences, suppresses spinner lines, adds semantic prefixes (`⟨tool⟩`, `⟨error⟩`, `⟨code⟩`) so your screen reader knows what it's reading
- **TTS mode** (`--tts`) — speaks agent output through the native TTS engine (macOS `say`, Linux `espeak-ng`/`spd-say`, Windows SAPI5) with a priority queue: errors interrupt, spinners are suppressed
- **High contrast mode** — replaces Unicode decoration with plain ASCII for low-vision users
- **Native SR bus** — posts announcements directly to VoiceOver (macOS), Orca/AT-SPI2 (Linux), or a named pipe for NVDA/JAWS scripts (Windows)
- **Transparent pass-through** — sighted users see exactly what they'd see without ariavox; SR/TTS features are opt-in
- **YAML config** — persistent preferences so you don't need flags every time

## Installation

### One-line installer (macOS / Linux)

```sh
curl -fsSL https://raw.githubusercontent.com/AgusRdz/ariavox/main/install.sh | sh
```

The installer verifies the SHA-256 checksum and, if `cosign` is installed, verifies the release signature.

### Homebrew (macOS / Linux)

```sh
brew install AgusRdz/tap/ariavox
```

### Manual

Download the binary for your platform from the [releases page](https://github.com/AgusRdz/ariavox/releases/latest):

| Platform | Binary |
|---|---|
| macOS (Apple Silicon + Intel) | `ariavox-darwin-universal` |
| macOS (Intel only) | `ariavox-darwin-amd64` |
| macOS (Apple Silicon only) | `ariavox-darwin-arm64` |
| Linux x86-64 | `ariavox-linux-amd64` |
| Linux ARM64 | `ariavox-linux-arm64` |
| Windows x86-64 | `ariavox-windows-amd64.exe` |

```sh
# macOS / Linux
chmod +x ariavox-darwin-universal
sudo mv ariavox-darwin-universal /usr/local/bin/ariavox

# verify (optional, requires cosign)
cosign verify-blob \
  --bundle checksums.txt.bundle \
  --certificate-identity-regexp "https://github.com/AgusRdz/ariavox/" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  checksums.txt
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/AgusRdz/ariavox/main/install.ps1 | iex
```

## Usage

### Basic — transparent pass-through

```sh
ariavox run -- claude
```

The agent runs exactly as normal. No accessibility features are active unless you add a flag or set them in config.

### Screen reader mode

```sh
ariavox run --sr -- claude
ariavox run --sr -- claude --dangerously-skip-permissions
```

- Strips all ANSI color and cursor codes
- Suppresses spinner lines (⠋ Thinking…)
- Adds semantic prefixes before tool use, errors, and code blocks
- Posts announcements to the native SR bus (VoiceOver / Orca / NVDA)

Enable permanently via config or environment variable:

```sh
ariavox config set screen_reader.enabled true
# or
ARIAVOX_SR=1 ariavox run -- claude
```

### TTS mode

```sh
ariavox run --tts -- claude
```

Speaks agent output through the system TTS engine. Priority rules:

- **Errors** and **tool use** interrupt current speech immediately
- **Regular text** is queued and spoken in order
- **Spinners** are suppressed

```sh
ariavox config set tts.enabled true
ariavox config set tts.rate 65       # 0-100, default 50
ariavox config set tts.voice "Ava"   # macOS voice name, empty = system default
```

### Both at once

```sh
ariavox run --sr --tts -- claude
```

### High contrast

```sh
# Via config
ariavox config set renderer.high_contrast true

# Via environment (respects NO_COLOR convention)
NO_COLOR=1 ariavox run --sr -- claude
```

Replaces Unicode decoration (`●`, `⎿`, Braille spinners) with plain ASCII (`*`, `->`) and uses bracket prefixes (`[tool]`, `[error]`, `[code]`).

## Configuration

```sh
ariavox config show          # print active config with file path
ariavox config edit          # open in $VISUAL / $EDITOR
ariavox config set <key> <value>
ariavox config path          # print config file location
```

**Config file location:**
- macOS / Linux: `~/.config/ariavox/ariavox.yaml`
- Windows: `%APPDATA%\ariavox\ariavox.yaml`

**Full reference:**

```yaml
tts:
  enabled: false
  rate: 50          # 0-100
  volume: 80        # 0-100
  voice: ""         # empty = system default
  priority_filter:  # which event kinds are spoken
    - task_end
    - error
    - tool_use

screen_reader:
  enabled: false
  strip_ansi: true
  suppress_spinners: true

renderer:
  separators: true         # blank line between different event types
  semantic_prefixes: true  # ⟨tool⟩, ⟨error⟩, ⟨code⟩ prefixes
  high_contrast: false     # ASCII-only decoration
  respect_no_color: true   # honor NO_COLOR env var

agent:
  command: ["claude"]
  args: []
```

### Priority filter values

`tts.priority_filter` accepts: `text`, `code`, `tool_use`, `tool_result`, `error`, `task_start`, `task_end`, `spinner`, `clear`.

## Environment variables

| Variable | Effect |
|---|---|
| `ARIAVOX_SR=1` | Enable screen reader mode (same as `--sr`) |
| `NO_COLOR=1` | Enable high contrast / ASCII mode |

## System requirements

Run `ariavox doctor` to check your system:

```
ariavox doctor — platform: darwin/arm64

dependencies:
  ok      say (TTS macOS) (/usr/bin/say)
  --      curl (optional, not found)

config:
  ok      /Users/you/.config/ariavox/ariavox.yaml

environment:
  ARIAVOX_SR not set
  NO_COLOR not set

all checks passed
```

**TTS backends by platform:**

| Platform | Backend | Install |
|---|---|---|
| macOS | `say` | built-in |
| Linux | `spd-say` (preferred) | `apt install speech-dispatcher` |
| Linux | `espeak-ng` (fallback) | `apt install espeak-ng` |
| Windows | PowerShell + SAPI5 | built-in |

**Screen reader bus by platform:**

| Platform | Implementation |
|---|---|
| macOS | NSAccessibility (`NSAccessibilityAnnouncementRequestedNotification`) |
| Linux | AT-SPI2 via D-Bus (`org.a11y.atspi.Event.Object.Announcement`) |
| Windows | Named pipe `\\.\pipe\ariavox-sr` (NVDA addon / JAWS script) |

## Building from source

Requires Go 1.22+. macOS builds need CGo for the VoiceOver integration.

```sh
# Run tests
make test

# Build for current platform (native, no Docker)
make install

# Build all platforms via Docker
make cross

# Check dependencies
make doctor
```

## Verifying releases

All releases are signed with [cosign](https://github.com/sigstore/cosign) using keyless signing (Sigstore). The signing identity is the GitHub Actions workflow in this repository.

```sh
cosign verify-blob \
  --bundle checksums.txt.bundle \
  --certificate-identity-regexp "https://github.com/AgusRdz/ariavox/" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  checksums.txt
```

Build provenance (SLSA3) is available via GitHub's attestation API:

```sh
gh attestation verify ariavox-darwin-universal \
  --owner AgusRdz
```

## License

MIT

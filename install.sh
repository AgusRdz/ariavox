#!/bin/sh
set -e

REPO="AgusRdz/ariavox"

# --- OS detection ---
OS="$(uname -s)"
case "$OS" in
  Linux*)  OS="linux" ;;
  Darwin*) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
  *) echo "unsupported OS: $OS" >&2; exit 1 ;;
esac

# --- Architecture detection ---
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# --- Install directory ---
if [ -z "$ARIAVOX_INSTALL_DIR" ]; then
  if [ "$OS" = "windows" ]; then
    INSTALL_DIR="$(cygpath "$LOCALAPPDATA/Programs/ariavox" 2>/dev/null || echo "$HOME/AppData/Local/Programs/ariavox")"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
else
  INSTALL_DIR="$ARIAVOX_INSTALL_DIR"
fi

EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi

# macOS: prefer universal binary
if [ "$OS" = "darwin" ]; then
  BINARY="ariavox-darwin-universal"
else
  BINARY="ariavox-${OS}-${ARCH}${EXT}"
fi

# --- Version ---
if [ -z "$ARIAVOX_VERSION" ]; then
  ARIAVOX_VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//')
fi

if [ -z "$ARIAVOX_VERSION" ]; then
  echo "failed to determine latest version" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${ARIAVOX_VERSION}/${BINARY}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${ARIAVOX_VERSION}/checksums.txt"
BUNDLE_URL="https://github.com/${REPO}/releases/download/${ARIAVOX_VERSION}/checksums.txt.bundle"

echo "installing ariavox ${ARIAVOX_VERSION} (${OS}/${ARCH})..."

# --- Download to temp dir first ---
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

TMP_BINARY="${TMP_DIR}/${BINARY}"
curl -fsSL "$URL" -o "$TMP_BINARY"

# --- Checksum verification ---
curl -fsSL "$CHECKSUMS_URL" -o "${TMP_DIR}/checksums.txt"

# Extract expected hash for this binary
EXPECTED=$(grep "  ${BINARY}$\| \*${BINARY}$\| ${BINARY}$" "${TMP_DIR}/checksums.txt" | awk '{print $1}')

if [ -z "$EXPECTED" ]; then
  echo "warning: ${BINARY} not found in checksums.txt — skipping checksum verification" >&2
else
  # Compute actual hash — use shasum on macOS, sha256sum on Linux
  if command -v shasum >/dev/null 2>&1; then
    ACTUAL=$(shasum -a 256 "$TMP_BINARY" | awk '{print $1}')
  elif command -v sha256sum >/dev/null 2>&1; then
    ACTUAL=$(sha256sum "$TMP_BINARY" | awk '{print $1}')
  else
    echo "warning: no sha256 tool found — skipping checksum verification" >&2
    ACTUAL="$EXPECTED"
  fi

  if [ "$ACTUAL" != "$EXPECTED" ]; then
    echo "checksum mismatch — aborting" >&2
    echo "  expected: $EXPECTED" >&2
    echo "  actual:   $ACTUAL" >&2
    exit 1
  fi
  echo "checksum verified"
fi

# --- Optional cosign verification ---
if command -v cosign >/dev/null 2>&1; then
  curl -fsSL "$BUNDLE_URL" -o "${TMP_DIR}/checksums.txt.bundle"
  if cosign verify-blob \
      --bundle "${TMP_DIR}/checksums.txt.bundle" \
      --certificate-identity-regexp "https://github.com/${REPO}/" \
      --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
      "${TMP_DIR}/checksums.txt" 2>/dev/null; then
    echo "cosign signature verified"
  else
    echo "warning: cosign verification failed" >&2
  fi
fi

# --- Install ---
mkdir -p "$INSTALL_DIR"
cp "$TMP_BINARY" "${INSTALL_DIR}/ariavox${EXT}"
chmod +x "${INSTALL_DIR}/ariavox${EXT}"
echo "installed ariavox to ${INSTALL_DIR}/ariavox${EXT}"
echo ""

# --- PATH registration ---
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    if [ "$OS" = "windows" ]; then
      WIN_DIR=$(cygpath -w "$INSTALL_DIR" 2>/dev/null || echo "$INSTALL_DIR")
      powershell.exe -NoProfile -Command \
        "\$p = [Environment]::GetEnvironmentVariable('Path', 'User'); \$d = '${WIN_DIR}'.TrimEnd('\\'); if ((\$p -split ';' | ForEach-Object { \$_.TrimEnd('\\') }) -notcontains \$d) { [Environment]::SetEnvironmentVariable('Path', \"\$d;\$p\", 'User'); Write-Host \"Added \$d to User PATH\" }"
      export PATH="${INSTALL_DIR}:$PATH"
    else
      SHELL_NAME="$(basename "${SHELL:-}")"
      case "$SHELL_NAME" in
        zsh)  SHELL_RC="$HOME/.zshrc" ;;
        bash) SHELL_RC="$HOME/.bashrc" ;;
        *)    SHELL_RC="" ;;
      esac

      PATH_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""

      if [ -n "$SHELL_RC" ]; then
        if ! grep -qF "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
          printf '\n# ariavox\n%s\n' "$PATH_LINE" >> "$SHELL_RC"
          echo "added ${INSTALL_DIR} to PATH in $SHELL_RC"
        fi
      else
        echo "NOTE: add this to your shell config:"
        echo "  $PATH_LINE"
      fi

      export PATH="${INSTALL_DIR}:$PATH"
      echo "ariavox is available in this shell session immediately"
      echo ""
    fi
    ;;
esac

echo "next steps:"
echo ""
echo "  ariavox run -- claude"
echo "  ariavox run --sr -- claude          # screen reader mode"
echo "  ariavox run --tts -- claude         # TTS mode"
echo "  ariavox config show"
echo "  ariavox doctor                      # verify system dependencies"
echo ""
echo "installation complete!"

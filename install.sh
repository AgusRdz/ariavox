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

BINARY="ariavox-${OS}-${ARCH}${EXT}"

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

echo "installing ariavox ${ARIAVOX_VERSION} (${OS}/${ARCH})..."

mkdir -p "$INSTALL_DIR"
curl -fsSL "$URL" -o "${INSTALL_DIR}/ariavox${EXT}"
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

      # Make available in the current shell session without restart
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

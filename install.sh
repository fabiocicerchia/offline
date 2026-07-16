#!/usr/bin/env bash
set -euo pipefail
# One-line installer for offline
# Usage: curl -fsSL https://raw.githubusercontent.com/fabiocicerchia/offline/main/install.sh | bash
#
# Clones (or updates) a checkout and builds+installs the binary via
# `make install` — see README for the equivalent manual steps.

REPO_DIR="${OFFLINE_HOME:-$HOME/.local/share/offline}"

if [ -d "$REPO_DIR/.git" ]; then
  git -C "$REPO_DIR" pull --ff-only
else
  git clone --depth 1 https://github.com/fabiocicerchia/offline "$REPO_DIR"
fi

make -C "$REPO_DIR" install
echo "offline installed. Run 'offline <program> [args...]'."

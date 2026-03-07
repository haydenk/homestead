#!/usr/bin/env bash
set -euo pipefail

# Add `mise activate` to shell rc files if not already present
for rc in "$HOME/.zshrc" "$HOME/.bashrc"; do
  if [ -f "$rc" ]; then
    shell=$(basename "$rc" rc | tr -d '.')
    activate_line='eval "$(mise activate '"$shell"')"'
    if ! grep -qF "mise activate" "$rc"; then
      echo "$activate_line" >> "$rc"
      echo "Added mise activate to $rc"
    fi
  fi
done

mise trust --yes
mise install

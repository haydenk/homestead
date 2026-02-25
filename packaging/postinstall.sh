#!/bin/sh
set -e

# Create system user if it doesn't exist
if ! id -u homestead >/dev/null 2>&1; then
  useradd --system --no-create-home --shell /usr/sbin/nologin homestead
fi

# Ensure the homestead user owns the config directory
chown -R homestead:homestead /etc/homestead

systemctl daemon-reload
systemctl enable homestead
systemctl start homestead || true

echo "Homestead installed."
echo "Edit /etc/homestead/config.toml then run: systemctl restart homestead"

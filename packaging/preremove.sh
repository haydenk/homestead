#!/bin/sh
set -e

systemctl stop    homestead 2>/dev/null || true
systemctl disable homestead 2>/dev/null || true

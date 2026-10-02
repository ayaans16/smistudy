#!/usr/bin/env bash
# Daily database backup with 14 days of history. Install with `crontab -e`:
#   15 4 * * * $HOME/smistudy/deploy/backup.sh >> $HOME/smistudy-backups/backup.log 2>&1
set -euo pipefail

out="$HOME/smistudy-backups/smistudy-$(date +%Y-%m-%d-%H%M).db"
"$HOME/smistudy-releases/current/smistudy-api" backup -db "$HOME/smistudy-data/smistudy.db" -out "$out"
chmod 600 "$out"
find "$HOME/smistudy-backups" -name 'smistudy-*.db' -mtime +14 -delete

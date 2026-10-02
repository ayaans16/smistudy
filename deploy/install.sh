#!/usr/bin/env bash
# One-time server setup for running smistudy under systemd with 2 API + 2 web instances.
#   sudo ./deploy/install.sh <linux-user>
set -euo pipefail

user="${1:?usage: sudo $0 <linux-user that owns the repo>}"
home=$(getent passwd "$user" | cut -d: -f6)
here="$(cd "$(dirname "$0")" && pwd)"

# Call the real node binary: the /snap/bin launcher is setuid and breaks under NoNewPrivileges.
node=/snap/node/current/bin/node
[ -x "$node" ] || node=$(command -v node)
[ -x "$node" ] || { echo "Node.js not found" >&2; exit 1; }

install -d -o "$user" -g "$user" -m 700 "$home/smistudy-data" "$home/smistudy-backups"
install -d -o "$user" -g "$user" -m 755 "$home/smistudy-releases"

# Carry over data from the single-user version (imported on first start).
for old in "$here/../backend/data/sessions.json" "$here/../backend/data/smistudy.db"; do
  name=$(basename "$old")
  if [ -f "$old" ] && [ ! -e "$home/smistudy-data/$name" ]; then
    install -o "$user" -g "$user" -m 600 "$old" "$home/smistudy-data/$name"
    echo "copied $old into $home/smistudy-data/"
  fi
done

if [ ! -f /etc/smistudy.env ]; then
  sed "s|__HOME__|$home|g" "$here/smistudy.env.example" > /etc/smistudy.env
  chmod 600 /etc/smistudy.env
  echo "created /etc/smistudy.env — fill in Google/Resend keys there"
fi

for unit in smistudy-api@.service smistudy-web@.service; do
  sed -e "s|__USER__|$user|g" -e "s|__HOME__|$home|g" -e "s|__NODE__|$node|g" \
    "$here/systemd/$unit" > "/etc/systemd/system/$unit"
done
systemctl daemon-reload
systemctl enable smistudy-api@8081 smistudy-api@8082 smistudy-web@3101 smistudy-web@3102

# Let the Caddy container (Docker network 172.18.0.0/16) reach the instances; nothing else can.
if command -v ufw >/dev/null; then
  ufw allow from 172.18.0.0/16 to any port 8081,8082,3101,3102 proto tcp
fi

echo "Installed. Next: run ./deploy/release.sh as $user to build and start everything."

#!/usr/bin/env bash
# Build the current checkout into a new release and roll it out one instance at a time,
# so the site stays up during deploys:  ./deploy/release.sh
set -euo pipefail

repo="$(cd "$(dirname "$0")/.." && pwd)"
releases="$HOME/smistudy-releases"
id="$(date +%Y%m%d-%H%M%S)-$(git -C "$repo" rev-parse --short HEAD)"
dir="$releases/$id"

echo "==> building release $id"
mkdir -p "$dir"
(cd "$repo/backend" && CGO_ENABLED=0 go build -trimpath -o "$dir/smistudy-api" .)
rsync -a --delete --exclude node_modules --exclude .next "$repo/frontend/" "$dir/frontend/"
(cd "$dir/frontend" && npm ci --no-audit --no-fund && SMISTUDY_API_URL=http://127.0.0.1:8081 npm run build)

echo "==> switching current -> $id"
ln -sfn "$dir" "$releases/current.tmp"
mv -T "$releases/current.tmp" "$releases/current"

wait_healthy() {
  for _ in $(seq 1 30); do
    curl -fsS -o /dev/null "$1" && return 0
    sleep 1
  done
  echo "!! $1 did not become healthy — check: journalctl -u $2 -n 50" >&2
  exit 1
}

# Restart one instance at a time; Caddy sends traffic to the other meanwhile.
for port in 8081 8082; do
  sudo systemctl restart "smistudy-api@$port"
  wait_healthy "http://127.0.0.1:$port/api/health" "smistudy-api@$port"
done
for port in 3101 3102; do
  sudo systemctl restart "smistudy-web@$port"
  wait_healthy "http://127.0.0.1:$port/login" "smistudy-web@$port"
done

# Keep the 5 newest releases for quick rollbacks.
ls -1dt "$releases"/*-*/ | tail -n +6 | while read -r old; do
  [ "$(readlink -f "$releases/current")/" = "$(readlink -f "$old")/" ] || rm -rf "$old"
done
echo "==> released $id"

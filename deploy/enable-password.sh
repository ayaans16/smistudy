#!/bin/sh
# Put the whole site behind a username/password (HTTP basic auth).
#   ./deploy/enable-password.sh <username>    enable, or add/replace a user
#   ./deploy/enable-password.sh --off         remove protection
set -eu

cd "$(dirname "$0")/nginx/auth"

if [ "${1:-}" = "--off" ]; then
  rm -f smistudy.conf .htpasswd
else
  user="${1:?usage: $0 <username> | --off}"
  printf "Password for %s: " "$user"
  stty -echo; read -r pass; stty echo; echo
  entry=$(printf '%s\n' "$pass" | docker run --rm -i httpd:2.4-alpine htpasswd -niB "$user")
  touch .htpasswd
  grep -v "^$user:" .htpasswd > .htpasswd.tmp || true
  printf '%s\n' "$entry" >> .htpasswd.tmp
  mv .htpasswd.tmp .htpasswd
  printf 'auth_basic "smistudy";\nauth_basic_user_file /etc/nginx/auth/.htpasswd;\n' > smistudy.conf
fi

docker compose exec nginx nginx -s reload 2>/dev/null && echo "nginx reloaded." || echo "Saved; applies next time nginx starts."

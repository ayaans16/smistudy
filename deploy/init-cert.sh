#!/bin/sh
# One-time: get the first Let's Encrypt certificate, then start the stack.
# Run from the repo root on the VPS after DNS points at it:  ./deploy/init-cert.sh
set -eu

cd "$(dirname "$0")/.."
[ -f .env ] || { echo "Missing .env — copy .env.example to .env and fill it in." >&2; exit 1; }
. ./.env
: "${DOMAIN:?set DOMAIN in .env}" "${EMAIL:?set EMAIL in .env}"

# Set a password before the site goes public (skip with NO_PASSWORD=1).
if [ "${NO_PASSWORD:-0}" != "1" ] && [ ! -f deploy/nginx/auth/smistudy.conf ]; then
  printf "Username for the site password: "
  read -r site_user
  ./deploy/enable-password.sh "$site_user"
fi

domains="-d $DOMAIN"
[ "${WWW:-1}" = "1" ] && domains="$domains -d www.$DOMAIN"

# nginx needs the certificate to start, so certbot briefly serves port 80 itself.
docker compose stop nginx 2>/dev/null || true
docker compose run --rm -p 80:80 --entrypoint certbot certbot \
  certonly --standalone --non-interactive --agree-tos -m "$EMAIL" $domains

docker compose up -d --build
echo "smistudy is up at https://$DOMAIN"

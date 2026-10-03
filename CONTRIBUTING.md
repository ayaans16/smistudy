# Contributing to smistudy

Thanks for helping make studying a little more fun! This guide covers running smistudy locally, how the code is organized, and how to get a change merged.

By taking part, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md). Please report security problems privately as described in [SECURITY.md](SECURITY.md), not in public issues.

## Ways to help

- **Bugs:** open an issue with steps to reproduce, what you expected, and what happened (screenshots help).
- **Ideas:** open an issue to discuss a feature before building it, so we can agree on the approach.
- **Code:** pick an issue, or fix something small and open a pull request.

## Project layout

| Part | Tech | Location |
| --- | --- | --- |
| API & logic | Go, SQLite | `backend/` |
| UI | Next.js 16, Tailwind CSS 4 | `frontend/` |
| Deployment | systemd, Caddy, shell scripts | `deploy/` |

The Go server stores everything in SQLite and does the calculations: the calendar grid, intensity levels, streaks and stats. Next.js forwards `/api/*` to it in development, so the browser only talks to one origin. In production, Caddy routes `/api/*` straight to the Go servers.

## Running it locally

You need **Go 1.26.6+** and **Node 20.9+**.

```bash
cd backend && go run .
```

```bash
cd frontend && npm install && npm run dev
```

Then open http://localhost:3000. Without a Resend key, sign-up confirmation and password-reset links are printed in the API's terminal output instead of being emailed, so copy them from there.

### Configuration

All optional for local development.

| Variable | Default | Purpose |
| --- | --- | --- |
| `SMISTUDY_ADDR` | `127.0.0.1:8080` | API listen address |
| `SMISTUDY_DB` | `data/smistudy.db` | SQLite database file |
| `SMISTUDY_DATA` | `data/sessions.json` | Old single-user JSON store; imported once on startup if present, then renamed to `.imported` |
| `SMISTUDY_PUBLIC_URL` | `http://localhost:3000` | The site's origin, used for email links, Google redirects, CSRF checks and secure cookies |
| `SMISTUDY_CLIENT_IP_HEADER` | (none) | Header with the real visitor IP for rate limiting, e.g. `X-Real-IP` behind the production Caddy config. Leave unset unless the API is only reachable through that proxy |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | (none) | Turn on "Continue with Google" |
| `RESEND_API_KEY` | (none) | Send emails through Resend. Without it, emails are printed to the API log |
| `SMISTUDY_MAIL_FROM` | `smistudy <noreply@smistudy.ca>` | Sender address for emails |
| `SMISTUDY_API_URL` | `http://localhost:8080` | Where Next.js proxies `/api` (frontend, read at build time) |

### Admin commands

```bash
./smistudy-api claim-legacy -email you@example.com   # attach pre-accounts sessions to an account
./smistudy-api backup -out backup.db                 # consistent snapshot of the live database
```

## Making a change

1. **Start from an up-to-date `main`** and create a branch. Never commit directly to `main`.
   ```bash
   git checkout main && git pull
   ```
   ```bash
   git checkout -b short-description-of-change
   ```
2. **Keep it focused.** One change per pull request is easier to review than several bundled together.
3. **Match the surrounding code:** naming, comment style and structure.
4. **Add or update tests** for backend behavior. Security-relevant code (auth, rate limits, data access) needs a test showing the protection works.
5. **Run the checks** CI will run:
   ```bash
   cd backend && gofmt -l . && go vet ./... && go test -race ./...
   ```
   ```bash
   cd frontend && npm run lint && npx next typegen && npx tsc --noEmit && npm run build
   ```
6. **Open a pull request** against `main` that explains what changed, why, and how you tested it. CI must pass before merging.

### Database changes

Add a new numbered file in `backend/migrations/` (e.g. `0004_something.sql`). Never edit a migration that has already been merged, since it has already run on the live database. Migrations run automatically, in order, when the API starts.

### Things to keep in mind

- **Privacy:** don't log passwords, tokens, emails or IP addresses, and don't add analytics or tracking. If a change collects new personal data or shares it with a new service, the [Privacy Policy](https://smistudy.ca/privacy) has to be updated in the same pull request.
- **Every query uses bound parameters.** Never build SQL from user input.
- **Every study-data query is scoped to the signed-in user.**
- **Original art only:** the mascots are original drawings. Don't add official Smiski or Sonny Angel artwork, photos or logos.

## Security model

- **Passwords** are hashed with Argon2id. Login sessions are random tokens in `HttpOnly`, `Secure`, `SameSite=Lax` cookies (`__Host-` prefixed over HTTPS), and only a SHA-256 of each token is stored.
- **Email verification** is required before password login. Reset links are single-use and expire after 1 hour, and a reset signs out every other device.
- **Brute force:** per-IP and per-account rate limits on login, plus an account lockout with exponential backoff after 5 wrong passwords. Signup, password reset and email sending have their own limits.
- **Account enumeration:** signup and "forgot password" respond the same whether or not an email is registered, and failed logins take the same time either way.
- **CSRF:** Go's `CrossOriginProtection` rejects cross-site writes, on top of `SameSite` cookies.
- **SQL injection:** every query uses bound parameters.
- **Google sign-in** uses state and PKCE, and only accepts Google-verified emails. Linking Google to an account whose email was never verified drops that account's password and sessions, to block account pre-hijacking.
- **Other:** request size limits, strict JSON decoding, server timeouts, security headers (API and website), sandboxed systemd services, and generic 500 errors (details go to the log only).

## API

Everything except `/api/health`, `/api/auth/*` and public profiles needs a signed-in session and only touches that user's data.

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/auth/signup` | `{ "email", "username", "password", "acceptTerms": true }`, then emails a verification link |
| POST | `/api/auth/verify` | `{ "token" }` from the email; signs you in |
| POST | `/api/auth/login` / `/api/auth/logout` | Password sign-in / sign-out |
| POST | `/api/auth/forgot` / `/api/auth/reset` | Email a reset link / `{ "token", "password" }` |
| POST | `/api/auth/resend-verification` | `{ "email" }` |
| GET | `/api/auth/google/start` | Redirects to Google sign-in |
| GET | `/api/auth/providers` | Which sign-in methods are enabled |
| GET / PATCH / DELETE | `/api/me` | Your account; update `username`, `displayName`, `profilePublic`; delete account |
| POST | `/api/me/password` | `{ "current", "new" }`; signs out other devices |
| GET | `/api/me/export` | Download all your data as JSON |
| GET | `/api/contributions?filter=last\|2026&today=YYYY-MM-DD` | Calendar grid for the graph |
| GET | `/api/years?today=…` | Years available in the filter |
| GET | `/api/stats?today=…` | Today, week, streaks, totals |
| GET | `/api/sessions?date=YYYY-MM-DD` | Sessions on one day |
| POST | `/api/sessions` | `{ "date", "minutes", "kind": "pomodoro"\|"manual", "note" }` |
| DELETE | `/api/sessions/{id}` | Remove a session |
| GET | `/api/users/{username}` (`/contributions`, `/years`) | Public profile and graph. No sign-in needed; returns 404 unless the user made their profile public |
| GET | `/api/users/{username}/card.svg?theme=light\|dark` | Embeddable SVG stats card (mini heatmap, hours, streaks). Private and unknown users get the same placeholder card; cached for 30 minutes |

The client sends `today` as its own local date, so days roll over at the user's midnight, not the server's. Graph intensity levels: none, under 1h, 1–2h, 2–4h, 4h+.

## License

smistudy is licensed under the [GNU Affero General Public License v3.0](LICENSE.md). By contributing, you agree that your contributions are licensed under the same terms.

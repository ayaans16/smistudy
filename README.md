# smistudy

A cozy, smiski & sonny angel–themed study page.

- **Study graph**: a GitHub-style contribution graph of the hours you study each day. The year filter works like GitHub's: "Last year" (rolling 12 months) or any year you have data for. Hover a square for details, or click one to see and edit that day's sessions.
- **Pomodoro timer**: pick focus, short-break and long-break lengths, and how many rounds come before a long break. Finished focus rounds are logged automatically. "Stop & log" saves a focus round you end early.
- **Stats**: today, this week, current and longest streak, all-time hours, and your average per study day.
- **Light/dark mode**: light by default. In dark mode ("lights off") the smiski glows.

## Stack

| Part | Tech | Location |
| --- | --- | --- |
| API & logic | Go, SQLite | `backend/` |
| UI | Next.js 16, Tailwind CSS 4 | `frontend/` |

The Go server stores sessions in SQLite and does the calculations: the calendar grid, intensity levels, streaks and stats. Next.js forwards `/api/*` to it, so the browser only talks to one origin.

## Running it

You need Go 1.22+ and Node 20+.

```bash
cd backend && go run .
```

```bash
cd frontend && npm install && npm run dev
```

Then open http://localhost:3000.

### Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `SMISTUDY_ADDR` | `127.0.0.1:8080` | API listen address |
| `SMISTUDY_DB` | `data/smistudy.db` | SQLite database file |
| `SMISTUDY_DATA` | `data/sessions.json` | Old JSON store; imported once on startup if present, then renamed to `.imported` |
| `SMISTUDY_PUBLIC_URL` | `http://localhost:3000` | The site's origin, used for email links, Google redirects, CSRF checks and secure cookies |
| `SMISTUDY_CLIENT_IP_HEADER` | (none) | Header with the real visitor IP for rate limiting, e.g. `CF-Connecting-IP` behind Cloudflare. Leave unset unless the API is only reachable through that proxy |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | (none) | Turn on "Sign in with Google" |
| `RESEND_API_KEY` | (none) | Send emails through Resend. Without it, emails (verification and reset links) are printed to the API log |
| `SMISTUDY_MAIL_FROM` | `smistudy <noreply@smistudy.ca>` | Sender address for emails |
| `SMISTUDY_API_URL` | `http://localhost:8080` | Where Next.js proxies `/api` (frontend, read at build time) |

Moving from the single-user version: old sessions have no owner after the upgrade. Sign up, then attach them to your account:

```bash
./smistudy-api claim-legacy -email you@example.com
```

## Security

- **Passwords** are hashed with Argon2id. Login sessions are random tokens in `HttpOnly`, `Secure`, `SameSite=Lax` cookies (`__Host-` prefixed over HTTPS), and only a SHA-256 of each token is stored.
- **Email verification** is required before password login. Reset links are single-use and expire after 1 hour, and a reset signs out every other device.
- **Brute force**: per-IP and per-account rate limits on login, plus an account lockout with exponential backoff after 5 wrong passwords. Signup, password reset and email sending have their own limits.
- **Account enumeration**: signup and "forgot password" respond the same whether or not an email is registered, and failed logins take the same time either way.
- **CSRF**: Go's `CrossOriginProtection` rejects cross-site writes, on top of `SameSite` cookies.
- **SQL injection**: every query uses bound parameters.
- **Google sign-in** uses state and PKCE, and only accepts Google-verified emails. Linking Google to an account whose email was never verified drops that account's password and sessions, to block account pre-hijacking.
- **Other**: request size limits, strict JSON decoding, server timeouts, security headers, and generic 500s (details go to the log only).

## API

Everything except `/api/health` and `/api/auth/*` needs a signed-in session and only touches that user's data.

| Method | Path | Description |
| --- | --- | --- |
| POST | `/api/auth/signup` | `{ "email", "username", "password" }`, then emails a verification link |
| POST | `/api/auth/verify` | `{ "token" }` from the email; signs you in |
| POST | `/api/auth/login` / `/api/auth/logout` | Password sign-in / sign-out |
| POST | `/api/auth/forgot` / `/api/auth/reset` | Email a reset link / `{ "token", "password" }` |
| POST | `/api/auth/resend-verification` | `{ "email" }` |
| GET | `/api/auth/google/start` | Redirects to Google sign-in |
| GET | `/api/auth/providers` | Which sign-in methods are enabled |
| GET / PATCH / DELETE | `/api/me` | Your account; update `username`, `displayName`, `profilePublic`; delete account |
| POST | `/api/me/password` | `{ "current", "new" }`; signs out other devices |
| GET | `/api/contributions?filter=last\|2026&today=YYYY-MM-DD` | Calendar grid for the graph |
| GET | `/api/years?today=…` | Years available in the filter |
| GET | `/api/stats?today=…` | Today, week, streaks, totals |
| GET | `/api/sessions?date=YYYY-MM-DD` | Sessions on one day |
| POST | `/api/sessions` | `{ "date", "minutes", "kind": "pomodoro"\|"manual", "note" }` |
| DELETE | `/api/sessions/{id}` | Remove a session |

The client sends `today` as its own local date, so days roll over at your midnight, not the server's.

Intensity levels: none, under 1h, 1–2h, 2–4h, 4h+.

The mascots are original fan-style drawings, not official smiski or sonny angel artwork.

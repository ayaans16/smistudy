# smistudy

A cozy, smiski & sonny angel–themed study page.

- **Study graph**: a GitHub-style contribution graph of the hours you study each day. The year filter works like GitHub's: "Last year" (rolling 12 months) or any year you have data for. Hover a square for details, or click one to see and edit that day's sessions.
- **Pomodoro timer**: pick focus, short-break and long-break lengths, and how many rounds come before a long break. Finished focus rounds are logged automatically. "Stop & log" saves a focus round you end early.
- **Stats**: today, this week, current and longest streak, all-time hours, and your average per study day.
- **Light/dark mode**: light by default. In dark mode ("lights off") the smiski glows.

## Stack

| Part | Tech | Location |
| --- | --- | --- |
| API & logic | Go (standard library only) | `backend/` |
| UI | Next.js 16, Tailwind CSS 4 | `frontend/` |

The Go server stores sessions in a JSON file and does the calculations: the calendar grid, intensity levels, streaks and stats. Next.js forwards `/api/*` to it, so the browser only talks to one origin.

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
| `SMISTUDY_ADDR` | `:8080` | API listen address |
| `SMISTUDY_DATA` | `data/sessions.json` | Where sessions are stored |
| `SMISTUDY_ORIGIN` | `http://localhost:3000` | Allowed CORS origin, for direct API calls |
| `SMISTUDY_API_URL` | `http://localhost:8080` | Where Next.js proxies `/api` (frontend) |

## API

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/contributions?filter=last\|2026&today=YYYY-MM-DD` | Calendar grid for the graph |
| GET | `/api/years?today=…` | Years available in the filter |
| GET | `/api/stats?today=…` | Today, week, streaks, totals |
| GET | `/api/sessions?date=YYYY-MM-DD` | Sessions on one day |
| POST | `/api/sessions` | `{ "date", "minutes", "kind": "pomodoro"\|"manual", "note" }` |
| DELETE | `/api/sessions/{id}` | Remove a session |

The client sends `today` as its own local date, so days roll over at your midnight, not the server's.

Intensity levels: none, under 1h, 1–2h, 2–4h, 4h+.

The mascots are original fan-style drawings, not official smiski or sonny angel artwork.

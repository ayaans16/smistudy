"use client";

import { useEffect, useState } from "react";
import { api, type Session } from "@/lib/api";
import { formatLongDate, formatMinutes } from "@/lib/format";
import { AngelBuddy } from "./Mascots";

type Props = {
  date: string;
  today: string;
  refreshKey: number;
  onChange: () => void;
};

/** Sessions for one day plus a small form for logging time studied away from the timer. */
export default function DayActivity({ date, today, refreshKey, onChange }: Props) {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [minutes, setMinutes] = useState("30");
  const [note, setNote] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    api
      .sessions(date)
      .then((s) => !cancelled && setSessions(s))
      .catch((e: Error) => !cancelled && setError(e.message));
    return () => {
      cancelled = true;
    };
  }, [date, refreshKey]);

  async function add(e: React.FormEvent) {
    e.preventDefault();
    const n = Number(minutes);
    if (!Number.isInteger(n) || n < 1) return setError("Enter a whole number of minutes.");
    try {
      await api.logSession({ date, minutes: n, kind: "manual", note });
      setNote("");
      setError(null);
      onChange();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function remove(id: string) {
    try {
      await api.deleteSession(id);
      onChange();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  const total = sessions.reduce((sum, s) => sum + s.minutes, 0);

  return (
    <section className="rounded-3xl border border-line bg-card p-6 shadow-sm">
      <div className="flex items-baseline justify-between gap-2">
        <h2 className="text-lg font-bold">
          {date === today ? "Today" : formatLongDate(date, true)}
        </h2>
        <span className="text-sm font-semibold text-muted">{formatMinutes(total)} total</span>
      </div>

      {sessions.length === 0 ? (
        <div className="flex items-center gap-3 py-5 text-sm text-muted">
          <AngelBuddy className="h-12 w-12 shrink-0" />
          No study sessions yet. Start a pomodoro or log some time below.
        </div>
      ) : (
        <ul className="mt-3 divide-y divide-line">
          {sessions.map((s) => (
            <li key={s.id} className="group flex items-center gap-3 py-2 text-sm">
              <span
                className={`h-2.5 w-2.5 shrink-0 rounded-full ${s.kind === "pomodoro" ? "bg-accent" : "bg-blush"}`}
                title={s.kind}
              />
              <span className="w-16 font-bold tabular-nums">{formatMinutes(s.minutes)}</span>
              <span className="flex-1 truncate text-muted">
                {s.note || (s.kind === "pomodoro" ? "Pomodoro" : "Logged manually")} ·{" "}
                {new Date(s.createdAt).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}
              </span>
              <button
                onClick={() => remove(s.id)}
                aria-label="Delete session"
                className="rounded-md px-2 text-muted opacity-60 transition hover:text-red-500 group-hover:opacity-100"
              >
                ×
              </button>
            </li>
          ))}
        </ul>
      )}

      <form onSubmit={add} className="mt-4 flex flex-wrap gap-2 border-t border-line pt-4">
        <input
          type="number"
          min={1}
          max={1440}
          value={minutes}
          onChange={(e) => setMinutes(e.target.value)}
          aria-label="Minutes"
          className="w-20 rounded-full border border-line bg-bg px-3 py-1.5 text-sm outline-none focus:border-accent"
        />
        <input
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="What did you study? (optional)"
          maxLength={200}
          className="min-w-0 flex-1 rounded-full border border-line bg-bg px-3 py-1.5 text-sm outline-none focus:border-accent"
        />
        <button className="rounded-full bg-fg px-4 py-1.5 text-sm font-bold text-bg transition hover:opacity-90">
          Log time
        </button>
      </form>
      {error && <p className="mt-2 text-sm text-red-500">{error}</p>}
    </section>
  );
}

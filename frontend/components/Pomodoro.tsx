"use client";

import { useEffect, useRef, useState } from "react";
import { api } from "@/lib/api";
import { formatClock, localDateKey } from "@/lib/format";
import { useStoredState } from "@/lib/useStoredState";
import { AngelBuddy, SmiBuddy } from "./Mascots";

type Mode = "focus" | "short" | "long";
type Status = "idle" | "running" | "paused";

const MODES: { id: Mode; label: string; options: number[] }[] = [
  { id: "focus", label: "Focus", options: [15, 25, 45, 50, 60, 90] },
  { id: "short", label: "Short break", options: [3, 5, 10] },
  { id: "long", label: "Long break", options: [15, 20, 30] },
];

const DEFAULTS: Record<Mode, number> & { longEvery: number } = { focus: 25, short: 5, long: 15, longEvery: 4 };

function chime() {
  try {
    const ctx = new AudioContext();
    [660, 880, 990].forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = "sine";
      osc.frequency.value = freq;
      const t = ctx.currentTime + i * 0.18;
      gain.gain.setValueAtTime(0.0001, t);
      gain.gain.exponentialRampToValueAtTime(0.2, t + 0.02);
      gain.gain.exponentialRampToValueAtTime(0.0001, t + 0.5);
      osc.connect(gain).connect(ctx.destination);
      osc.start(t);
      osc.stop(t + 0.55);
    });
  } catch {}
}

export default function Pomodoro({ onLogged }: { onLogged: () => void }) {
  const [settings, setSettings] = useStoredState("smistudy-pomodoro", DEFAULTS);
  const [mode, setMode] = useState<Mode>("focus");
  const [status, setStatus] = useState<Status>("idle");
  const [endAt, setEndAt] = useState(0);
  const [pausedLeft, setPausedLeft] = useState(0);
  const [now, setNow] = useState(0);
  const [rounds, setRounds] = useState(0);
  const [toast, setToast] = useState<{ text: string; error?: boolean } | null>(null);
  const finishing = useRef(false);

  const totalMs = settings[mode] * 60_000;
  const leftMs = status === "running" ? Math.max(0, endAt - now) : status === "paused" ? pausedLeft : totalMs;
  const progress = 1 - leftMs / totalMs;

  async function log(minutes: number) {
    try {
      await api.logSession({ date: localDateKey(), minutes, kind: "pomodoro" });
      setToast({ text: `Logged ${minutes} min — nice work!` });
      onLogged();
    } catch (e) {
      setToast({ text: `Couldn't log session: ${(e as Error).message}`, error: true });
    }
  }

  function switchTo(next: Mode) {
    setMode(next);
    setStatus("idle");
  }

  function skip() {
    switchTo(mode === "focus" ? "short" : "focus");
  }

  function finish() {
    if (finishing.current) return;
    finishing.current = true;
    chime();
    if (mode === "focus") {
      const done = rounds + 1;
      setRounds(done);
      void log(settings.focus);
      switchTo(done % settings.longEvery === 0 ? "long" : "short");
    } else {
      switchTo("focus");
    }
  }

  function start() {
    const t = Date.now();
    finishing.current = false;
    setNow(t);
    setEndAt(t + (status === "paused" ? pausedLeft : totalMs));
    setStatus("running");
  }

  function pause() {
    setPausedLeft(Math.max(0, endAt - Date.now()));
    setStatus("paused");
  }

  function stopAndLog() {
    const elapsed = Math.floor((totalMs - leftMs) / 60_000);
    if (elapsed >= 1) void log(elapsed);
    else setToast({ text: "Less than a minute — nothing logged." });
    switchTo("focus");
  }

  // Timestamp-based ticking keeps time accurate even when the tab is throttled.
  useEffect(() => {
    if (status !== "running") return;
    const id = setInterval(() => {
      const t = Date.now();
      setNow(t);
      if (t >= endAt) finish();
    }, 250);
    return () => clearInterval(id);
  });

  useEffect(() => {
    const label = mode === "focus" ? "focus" : "break";
    document.title = status === "idle" ? "smistudy" : `${formatClock(leftMs)} · ${label} — smistudy`;
  }, [leftMs, mode, status]);

  useEffect(() => {
    if (!toast) return;
    const id = setTimeout(() => setToast(null), 3500);
    return () => clearTimeout(id);
  }, [toast]);

  const radius = 88;
  const circumference = 2 * Math.PI * radius;
  const isBreak = mode !== "focus";
  const active = status !== "idle";

  return (
    <section className="rounded-3xl border border-line bg-card p-6 shadow-sm">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-bold">Pomodoro</h2>
        <span className="text-sm text-muted">
          Round {(rounds % settings.longEvery) + 1} of {settings.longEvery}
        </span>
      </div>

      {/* mode tabs */}
      <div className="mt-4 grid grid-cols-3 gap-1 rounded-full bg-bg p-1 text-sm font-semibold">
        {MODES.map((m) => (
          <button
            key={m.id}
            onClick={() => switchTo(m.id)}
            className={`rounded-full px-3 py-1.5 transition ${
              mode === m.id
                ? m.id === "focus"
                  ? "bg-accent text-[#1f2a14] shadow-sm"
                  : "bg-blush text-[#3a2228] shadow-sm"
                : "text-muted hover:text-fg"
            }`}
          >
            {m.label}
          </button>
        ))}
      </div>

      {/* timer ring with mascot */}
      <div className="relative mx-auto mt-6 h-56 w-56">
        <svg viewBox="0 0 200 200" className="h-full w-full -rotate-90">
          <circle cx="100" cy="100" r={radius} fill="none" stroke="var(--line)" strokeWidth="10" />
          <circle
            cx="100"
            cy="100"
            r={radius}
            fill="none"
            stroke={isBreak ? "var(--blush)" : "var(--accent)"}
            strokeWidth="10"
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={circumference * (1 - progress)}
            className="transition-[stroke-dashoffset] duration-300 ease-linear"
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          {isBreak ? (
            <AngelBuddy className={`h-16 w-16 ${status === "running" ? "animate-bob" : ""}`} />
          ) : (
            <SmiBuddy className={`h-16 w-16 ${status === "running" ? "animate-bob" : ""}`} />
          )}
          <div className="mt-1 text-5xl font-extrabold tabular-nums tracking-tight">{formatClock(leftMs)}</div>
          <div className="text-xs font-semibold uppercase tracking-widest text-muted">
            {isBreak ? "rest a little" : status === "running" ? "studying" : "ready?"}
          </div>
        </div>
      </div>

      {/* controls */}
      <div className="mt-6 flex items-center justify-center gap-2">
        {status === "running" ? (
          <button onClick={pause} className="min-w-28 rounded-full bg-fg px-6 py-2.5 font-bold text-bg transition hover:opacity-90">
            Pause
          </button>
        ) : (
          <button
            onClick={start}
            className={`min-w-28 rounded-full px-6 py-2.5 font-bold shadow-sm transition hover:brightness-105 ${
              isBreak ? "bg-blush text-[#3a2228]" : "bg-accent text-[#1f2a14]"
            }`}
          >
            {status === "paused" ? "Resume" : "Start"}
          </button>
        )}
        {active && (
          <button onClick={() => switchTo(mode)} className="rounded-full border border-line px-4 py-2.5 font-semibold text-muted transition hover:text-fg">
            Reset
          </button>
        )}
        {active && mode === "focus" ? (
          <button onClick={stopAndLog} className="rounded-full border border-line px-4 py-2.5 font-semibold text-muted transition hover:text-fg">
            Stop &amp; log
          </button>
        ) : (
          <button onClick={skip} className="rounded-full border border-line px-4 py-2.5 font-semibold text-muted transition hover:text-fg" title="Skip to next">
            Skip
          </button>
        )}
      </div>

      {/* clickable duration options */}
      <div className="mt-6 space-y-3 border-t border-line pt-5">
        {MODES.map((m) => {
          const locked = active && mode === m.id;
          return (
            <div key={m.id} className="flex flex-wrap items-center gap-2">
              <span className="w-24 text-sm font-semibold text-muted">{m.label}</span>
              {m.options.map((min) => {
                const selected = settings[m.id] === min;
                return (
                  <button
                    key={min}
                    disabled={locked}
                    onClick={() => setSettings({ ...settings, [m.id]: min })}
                    className={`rounded-full border px-3 py-1 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-40 ${
                      selected
                        ? m.id === "focus"
                          ? "border-accent bg-accent-soft text-accent-strong"
                          : "border-blush bg-blush-soft text-fg"
                        : "border-line text-muted hover:border-muted hover:text-fg"
                    }`}
                  >
                    {min}m
                  </button>
                );
              })}
            </div>
          );
        })}
        <div className="flex flex-wrap items-center gap-2">
          <span className="w-24 text-sm font-semibold text-muted">Long every</span>
          {[2, 3, 4, 5].map((n) => (
            <button
              key={n}
              onClick={() => setSettings({ ...settings, longEvery: n })}
              className={`rounded-full border px-3 py-1 text-sm font-semibold transition ${
                settings.longEvery === n ? "border-blush bg-blush-soft text-fg" : "border-line text-muted hover:border-muted hover:text-fg"
              }`}
            >
              {n} rounds
            </button>
          ))}
        </div>
      </div>

      <div aria-live="polite" className="h-6 pt-2 text-center text-sm font-semibold">
        {toast && <span className={toast.error ? "text-red-500" : "text-accent-strong"}>{toast.text}</span>}
      </div>
    </section>
  );
}

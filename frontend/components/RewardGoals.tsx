"use client";

import { useEffect, useState } from "react";
import { goals as goalApi, type RewardGoal } from "@/lib/api";
import { formatHours, formatLongDate, localDateKey } from "@/lib/format";
import { errorMessage } from "./AuthCard";
import { AngelBuddy } from "./Mascots";

const PRESET_HOURS = [5, 10, 20, 50];

function GiftIcon({ className = "" }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="3" y="8" width="18" height="4" rx="1" />
      <path d="M12 8v13M19 12v7a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2v-7" />
      <path d="M7.5 8a2.5 2.5 0 0 1 0-5C10 3 12 8 12 8s2-5 4.5-5a2.5 2.5 0 0 1 0 5" />
    </svg>
  );
}

/** Self-set rewards unlocked by studying a number of hours. Refreshes when sessions are logged. */
export default function RewardGoals({ refreshKey }: { refreshKey: number }) {
  const [items, setItems] = useState<RewardGoal[] | null>(null);
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    goalApi
      .list()
      .then((list) => !cancelled && setItems(list))
      .catch((e) => !cancelled && setError(errorMessage(e)));
    return () => {
      cancelled = true;
    };
  }, [refreshKey]);

  async function run(action: () => Promise<unknown>) {
    setError(null);
    try {
      await action();
      setItems(await goalApi.list());
    } catch (e) {
      setError(errorMessage(e));
    }
  }

  const active = items?.filter((g) => !g.claimedAt) ?? [];
  const claimed = items?.filter((g) => g.claimedAt) ?? [];

  return (
    <section className="rounded-3xl border border-line bg-card p-6 shadow-sm">
      <div className="flex items-center justify-between gap-2">
        <h2 className="flex items-center gap-2 text-lg font-bold">
          <GiftIcon className="h-5 w-5 text-accent-strong" />
          Rewards
        </h2>
        {!adding && (
          <button onClick={() => setAdding(true)} className="rounded-full border border-line px-3 py-1 text-sm font-bold transition hover:border-accent">
            + New goal
          </button>
        )}
      </div>

      {adding && (
        <NewGoalForm
          onCancel={() => setAdding(false)}
          onCreate={(b) =>
            run(async () => {
              await goalApi.add(b);
              setAdding(false);
            })
          }
        />
      )}

      {error && <p className="mt-3 text-sm text-red-500">{error}</p>}

      {items && active.length === 0 && !adding && (
        <div className="mt-3 flex items-center gap-3 text-sm text-muted">
          <AngelBuddy className="h-12 w-12 shrink-0" />
          Promise yourself a treat (bubble tea, a new blind box, a day off) for hitting a study goal.
        </div>
      )}

      <ul className="mt-4 space-y-3">
        {active.map((g) => (
          <GoalRow key={g.id} goal={g} onClaim={() => run(() => goalApi.claim(g.id))} onDelete={() => run(() => goalApi.remove(g.id))} />
        ))}
      </ul>

      {claimed.length > 0 && (
        <details className="mt-4 text-sm">
          <summary className="cursor-pointer font-semibold text-muted hover:text-fg">Claimed rewards ({claimed.length})</summary>
          <ul className="mt-2 space-y-1">
            {claimed.map((g) => (
              <li key={g.id} className="group flex items-center gap-2">
                <span className="flex-1 truncate">🎁 {g.reward}</span>
                <span className="text-xs text-muted">{formatHours(g.targetMinutes)}h</span>
                <button
                  onClick={() => run(() => goalApi.remove(g.id))}
                  aria-label={`Delete "${g.reward}"`}
                  className="px-1 text-muted opacity-100 hover:text-red-500 sm:opacity-0 sm:group-hover:opacity-100 sm:focus:opacity-100"
                >
                  ×
                </button>
              </li>
            ))}
          </ul>
        </details>
      )}
    </section>
  );
}

function GoalRow({ goal, onClaim, onDelete }: { goal: RewardGoal; onClaim: () => void; onDelete: () => void }) {
  const reached = Boolean(goal.reachedOn);
  const pct = Math.min(100, Math.round((goal.minutes / goal.targetMinutes) * 100));

  return (
    <li className={`group rounded-2xl border p-4 transition ${reached ? "border-blush bg-blush-soft" : "border-line"}`}>
      <div className="flex items-start gap-2">
        <div className="min-w-0 flex-1">
          <div className="break-words font-bold">{goal.reward}</div>
          <div className="mt-0.5 text-xs text-muted">
            {formatHours(Math.min(goal.minutes, goal.targetMinutes))}h of {formatHours(goal.targetMinutes)}h · since{" "}
            {formatLongDate(goal.startDate)}
          </div>
        </div>
        <button
          onClick={onDelete}
          aria-label={`Delete "${goal.reward}"`}
          className="shrink-0 px-1 text-muted opacity-100 transition hover:text-red-500 sm:opacity-0 sm:group-hover:opacity-100 sm:focus:opacity-100"
        >
          ×
        </button>
      </div>

      <div
        className="mt-3 h-2.5 overflow-hidden rounded-full bg-line"
        role="progressbar"
        aria-valuenow={pct}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-label={`${goal.reward} progress`}
      >
        <div className={`h-full rounded-full transition-all duration-500 ${reached ? "bg-blush" : "bg-accent"}`} style={{ width: `${pct}%` }} />
      </div>

      {reached ? (
        <div className="mt-3 flex flex-wrap items-center justify-between gap-2">
          <span className="text-sm font-bold">🎉 Unlocked on {formatLongDate(goal.reachedOn!)}!</span>
          <button onClick={onClaim} className="rounded-full bg-blush px-4 py-1.5 text-sm font-bold text-[#3a2228] shadow-sm transition hover:brightness-105">
            Claim reward
          </button>
        </div>
      ) : (
        <div className="mt-1.5 text-right text-xs font-semibold text-muted">
          {pct}% · {formatHours(goal.targetMinutes - goal.minutes)}h to go
        </div>
      )}
    </li>
  );
}

function NewGoalForm({
  onCreate,
  onCancel,
}: {
  onCreate: (b: { reward: string; targetHours: number; startDate: string }) => Promise<void>;
  onCancel: () => void;
}) {
  const [reward, setReward] = useState("");
  const [hours, setHours] = useState("10");
  const [busy, setBusy] = useState(false);
  const target = Number(hours);
  const valid = reward.trim() !== "" && target >= 0.5 && target <= 1000;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!valid) return;
    setBusy(true);
    await onCreate({ reward: reward.trim(), targetHours: target, startDate: localDateKey() });
    setBusy(false);
  }

  return (
    <form onSubmit={submit} className="mt-4 space-y-3 rounded-2xl bg-bg p-4">
      <label className="block">
        <span className="text-sm font-semibold">Reward</span>
        <input
          autoFocus
          value={reward}
          onChange={(e) => setReward(e.target.value)}
          maxLength={100}
          placeholder="e.g. a new smiski blind box"
          className="mt-1 w-full rounded-xl border border-line bg-card px-3 py-2 text-sm outline-none focus:border-accent"
        />
      </label>
      <div>
        <span className="text-sm font-semibold">After studying</span>
        <div className="mt-1 flex flex-wrap items-center gap-2">
          {PRESET_HOURS.map((h) => (
            <button
              type="button"
              key={h}
              onClick={() => setHours(String(h))}
              className={`rounded-full border px-3 py-1 text-sm font-semibold transition ${
                target === h ? "border-accent bg-accent-soft text-accent-strong" : "border-line text-muted hover:text-fg"
              }`}
            >
              {h}h
            </button>
          ))}
          <label className="flex items-center gap-1 text-sm text-muted">
            or
            <input
              type="number"
              min={0.5}
              max={1000}
              step={0.5}
              value={hours}
              onChange={(e) => setHours(e.target.value)}
              aria-label="Custom number of hours"
              className="w-20 rounded-full border border-line bg-card px-3 py-1 text-sm text-fg outline-none focus:border-accent"
            />
            hours
          </label>
        </div>
        <p className="mt-1 text-xs text-muted">Counts study time from today onward.</p>
      </div>
      <div className="flex gap-2">
        <button
          disabled={!valid || busy}
          className="rounded-full bg-accent px-4 py-1.5 text-sm font-bold text-[#1f2a14] transition hover:brightness-105 disabled:opacity-40"
        >
          {busy ? "Saving…" : "Set goal"}
        </button>
        <button type="button" onClick={onCancel} className="rounded-full px-3 py-1.5 text-sm font-semibold text-muted hover:text-fg">
          Cancel
        </button>
      </div>
    </form>
  );
}

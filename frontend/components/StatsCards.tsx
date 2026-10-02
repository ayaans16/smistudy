"use client";

import { useEffect, useState } from "react";
import { api, type Stats } from "@/lib/api";
import { formatHours, formatMinutes } from "@/lib/format";

export default function StatsCards({ today, refreshKey }: { today: string; refreshKey: number }) {
  const [stats, setStats] = useState<Stats | null>(null);

  useEffect(() => {
    let cancelled = false;
    api
      .stats(today)
      .then((s) => !cancelled && setStats(s))
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [today, refreshKey]);

  const cards = [
    { label: "Today", value: stats ? formatMinutes(stats.todayMinutes) : "–" },
    { label: "This week", value: stats ? formatMinutes(stats.weekMinutes) : "–" },
    { label: "Current streak", value: stats ? `${stats.currentStreak} ${stats.currentStreak === 1 ? "day" : "days"}` : "–" },
    { label: "Longest streak", value: stats ? `${stats.longestStreak} ${stats.longestStreak === 1 ? "day" : "days"}` : "–" },
    { label: "All time", value: stats ? `${formatHours(stats.totalMinutes)}h` : "–" },
    { label: "Avg / study day", value: stats ? formatMinutes(stats.dailyAverage) : "–" },
  ];

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
      {cards.map((c) => (
        <div key={c.label} className="rounded-2xl border border-line bg-card px-4 py-3 shadow-sm">
          <div className="text-xs font-semibold uppercase tracking-wide text-muted">{c.label}</div>
          <div className="mt-1 text-2xl font-extrabold tabular-nums">{c.value}</div>
        </div>
      ))}
    </div>
  );
}

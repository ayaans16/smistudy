"use client";

import { useState, useSyncExternalStore } from "react";
import DayActivity from "@/components/DayActivity";
import { AngelBuddy } from "@/components/Mascots";
import Pomodoro from "@/components/Pomodoro";
import SiteHeader from "@/components/SiteHeader";
import StatsCards from "@/components/StatsCards";
import StudyGraph from "@/components/StudyGraph";
import { localDateKey } from "@/lib/format";
import { useMe } from "@/lib/useMe";

// Re-check the local date every minute so the app rolls over at midnight.
function subscribeToClock(cb: () => void) {
  const id = setInterval(cb, 60_000);
  return () => clearInterval(id);
}

export default function Home() {
  const { me } = useMe({ required: true });
  const today = useSyncExternalStore(subscribeToClock, () => localDateKey(), () => "");
  const [refreshKey, setRefreshKey] = useState(0);
  const [selected, setSelected] = useState<string | null>(null);
  const refresh = () => setRefreshKey((k) => k + 1);

  return (
    <div className="mx-auto w-full max-w-6xl px-4 pb-16 sm:px-6">
      <SiteHeader me={me} />

      {today && me && (
        <main className="flex flex-col gap-8">
          <div className="grid gap-6 lg:grid-cols-[minmax(0,420px)_1fr]">
            <Pomodoro onLogged={refresh} />
            <div className="flex flex-col gap-6">
              <div className="relative overflow-hidden rounded-3xl border border-line bg-accent-soft p-6">
                <p className="max-w-[70%] text-xl font-extrabold leading-snug">
                  Hi {me.displayName || me.username}! Little by little, every session counts.
                </p>
                <p className="mt-1 max-w-[70%] text-sm text-muted">
                  Focus with the timer, take your breaks, and watch your garden grow.
                </p>
                <AngelBuddy className="absolute -bottom-2 right-4 h-24 w-24 animate-bob" />
              </div>
              <StatsCards today={today} refreshKey={refreshKey} />
              <DayActivity date={selected ?? today} today={today} refreshKey={refreshKey} onChange={refresh} />
            </div>
          </div>

          <StudyGraph today={today} refreshKey={refreshKey} selected={selected} onSelect={setSelected} />
        </main>
      )}

    </div>
  );
}

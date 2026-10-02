"use client";

import { useEffect, useState } from "react";
import { api, profiles, type Calendar, type Day } from "@/lib/api";
import { formatHours, formatLongDate, formatMinutes } from "@/lib/format";

const CELL = 11;
const GAP = 3;
const DAY_LABELS = ["", "Mon", "", "Wed", "", "Fri", ""];

type Props = {
  today: string;
  refreshKey?: number;
  selected?: string | null;
  /** Omit to make days non-clickable (e.g. on public profiles). */
  onSelect?: (date: string | null) => void;
  /** Show this user's public graph instead of the signed-in user's. */
  username?: string;
};

function legendTitle(level: number, t: number[]) {
  if (level === 0) return "No study time";
  if (level === 4) return `${t[3] / 60}h or more`;
  const lo = t[level - 1];
  const hi = t[level];
  return `${lo === 1 ? "Under" : `${formatMinutes(lo)} –`} ${formatMinutes(hi)}`;
}

export default function StudyGraph({ today, refreshKey = 0, selected = null, onSelect, username }: Props) {
  const [filter, setFilter] = useState("last");
  const [years, setYears] = useState<number[]>([]);
  const [cal, setCal] = useState<Calendar | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [hover, setHover] = useState<{ day: Day; x: number; y: number } | null>(null);

  useEffect(() => {
    let cancelled = false;
    const load = username
      ? Promise.all([profiles.contributions(username, filter, today), profiles.years(username, today)])
      : Promise.all([api.contributions(filter, today), api.years(today)]);
    load
      .then(([c, y]) => {
        if (cancelled) return;
        setCal(c);
        setYears(y);
        setError(null);
      })
      .catch((e: Error) => !cancelled && setError(e.message));
    return () => {
      cancelled = true;
    };
  }, [filter, today, refreshKey, username]);

  const heading = cal
    ? `${formatHours(cal.totalMinutes)} hours studied ${filter === "last" ? "in the last year" : `in ${filter}`}`
    : "Loading…";

  function chooseFilter(f: string) {
    setFilter(f);
    onSelect?.(null);
  }

  return (
    <section className="relative flex flex-col gap-4 lg:flex-row">
      <div className="min-w-0 flex-1">
        <h2 className="mb-2 text-base font-semibold">{heading}</h2>
        <div className="relative rounded-2xl border border-line bg-card p-4 shadow-sm">
          {error ? (
            <p className="py-10 text-center text-sm text-muted">
              Couldn&apos;t load the study graph ({error}).
            </p>
          ) : (
            <div className="overflow-x-auto pb-1">
              {cal && (
                <div className="inline-flex flex-col" style={{ minWidth: cal.weeks.length * (CELL + GAP) + 32 }}>
                  {/* month labels */}
                  <div className="relative ml-8 h-5 text-xs text-muted">
                    {cal.months.map((m) => (
                      <span key={`${m.name}-${m.week}`} className="absolute" style={{ left: m.week * (CELL + GAP) }}>
                        {m.name}
                      </span>
                    ))}
                  </div>
                  <div className="flex">
                    {/* weekday labels */}
                    <div className="flex w-8 flex-col text-[10px] text-muted" style={{ gap: GAP }}>
                      {DAY_LABELS.map((l, i) => (
                        <span key={i} className="leading-none" style={{ height: CELL, lineHeight: `${CELL}px` }}>
                          {l}
                        </span>
                      ))}
                    </div>
                    {/* grid */}
                    <div className="flex" style={{ gap: GAP }} onMouseLeave={() => setHover(null)}>
                      {cal.weeks.map((week, wi) => (
                        <div key={wi} className="flex flex-col" style={{ gap: GAP }}>
                          {week.map((day) =>
                            day.inRange ? (
                              <button
                                key={day.date}
                                aria-label={`${day.minutes ? formatMinutes(day.minutes) : "No study time"} on ${formatLongDate(day.date, true)}`}
                                onClick={onSelect && (() => onSelect(selected === day.date ? null : day.date))}
                                onMouseEnter={(e) => {
                                  const box = e.currentTarget.getBoundingClientRect();
                                  const parent = e.currentTarget.closest("section")!.getBoundingClientRect();
                                  setHover({ day, x: box.left - parent.left + CELL / 2, y: box.top - parent.top });
                                }}
                                className={`heat-${day.level} rounded-[3px] ${onSelect ? "" : "cursor-default"} outline-offset-1 transition-opacity ${
                                  selected === day.date ? "outline-2 outline-fg" : ""
                                } ${selected && selected !== day.date ? "opacity-40" : ""}`}
                                style={{ width: CELL, height: CELL, outlineStyle: selected === day.date ? "solid" : undefined }}
                              />
                            ) : (
                              <span key={day.date} style={{ width: CELL, height: CELL }} />
                            ),
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                  {/* footer */}
                  <div className="mt-3 flex items-center justify-between gap-4 text-xs text-muted">
                    <span>
                      {cal.activeDays} study {cal.activeDays === 1 ? "day" : "days"}
                    </span>
                    <div className="flex items-center gap-1">
                      <span className="mr-1">Less</span>
                      {[0, 1, 2, 3, 4].map((l) => (
                        <span
                          key={l}
                          title={legendTitle(l, cal.thresholds)}
                          className={`heat-${l} inline-block rounded-[3px]`}
                          style={{ width: CELL, height: CELL }}
                        />
                      ))}
                      <span className="ml-1">More</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* year filter, like GitHub's sidebar */}
      <nav className="flex gap-1 overflow-x-auto lg:mt-7 lg:w-28 lg:flex-col" aria-label="Filter by year">
        {["last", ...years.map(String)].map((f) => (
          <button
            key={f}
            onClick={() => chooseFilter(f)}
            className={`shrink-0 rounded-lg px-3 py-1.5 text-left text-sm transition ${
              filter === f ? "bg-accent font-bold text-[#1f2a14]" : "text-muted hover:bg-accent-soft hover:text-fg"
            }`}
          >
            {f === "last" ? "Last year" : f}
          </button>
        ))}
      </nav>

      {hover && (
        <div
          className="pointer-events-none absolute z-10 -translate-x-1/2 -translate-y-full whitespace-nowrap rounded-md bg-fg px-2 py-1 text-xs font-semibold text-bg shadow-lg"
          style={{ left: hover.x, top: hover.y - 6 }}
        >
          {hover.day.minutes ? `${formatMinutes(hover.day.minutes)} studied` : "No study time"} on{" "}
          {formatLongDate(hover.day.date)}.
        </div>
      )}
    </section>
  );
}

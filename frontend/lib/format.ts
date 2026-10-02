/** YYYY-MM-DD in the user's local timezone. */
export function localDateKey(d = new Date()): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/** 150 → "2h 30m", 45 → "45m", 0 → "0m" */
export function formatMinutes(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = Math.round(minutes % 60);
  if (h === 0) return `${m}m`;
  return m === 0 ? `${h}h` : `${h}h ${m}m`;
}

export function formatHours(minutes: number): string {
  const hours = minutes / 60;
  return hours.toLocaleString(undefined, { maximumFractionDigits: 1 });
}

function ordinal(n: number): string {
  const s = ["th", "st", "nd", "rd"];
  const v = n % 100;
  return n + (s[(v - 20) % 10] || s[v] || s[0]);
}

/** "2026-10-02" → "October 2nd" (GitHub tooltip style). */
export function formatLongDate(key: string, withYear = false): string {
  const [y, m, d] = key.split("-").map(Number);
  const month = new Date(y, m - 1, d).toLocaleString("en-US", { month: "long" });
  return `${month} ${ordinal(d)}${withYear ? `, ${y}` : ""}`;
}

/** mm:ss */
export function formatClock(ms: number): string {
  const total = Math.max(0, Math.ceil(ms / 1000));
  const m = Math.floor(total / 60);
  const s = total % 60;
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

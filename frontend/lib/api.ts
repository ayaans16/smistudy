export type Day = {
  date: string;
  minutes: number;
  sessions: number;
  level: 0 | 1 | 2 | 3 | 4;
  inRange: boolean;
};

export type Calendar = {
  from: string;
  to: string;
  totalMinutes: number;
  activeDays: number;
  weeks: Day[][];
  months: { name: string; week: number }[];
  thresholds: number[];
};

export type Stats = {
  todayMinutes: number;
  weekMinutes: number;
  totalMinutes: number;
  currentStreak: number;
  longestStreak: number;
  bestDay?: string;
  bestDayMinutes: number;
  dailyAverage: number;
};

export type Session = {
  id: string;
  date: string;
  minutes: number;
  kind: "pomodoro" | "manual";
  note?: string;
  createdAt: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
    cache: "no-store",
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? `Request failed (${res.status})`);
  }
  return res.status === 204 ? (undefined as T) : res.json();
}

export const api = {
  contributions: (filter: string, today: string) =>
    request<Calendar>(`/contributions?filter=${filter}&today=${today}`),
  stats: (today: string) => request<Stats>(`/stats?today=${today}`),
  years: (today: string) => request<number[]>(`/years?today=${today}`),
  sessions: (date: string) => request<Session[]>(`/sessions?date=${date}`),
  logSession: (s: Pick<Session, "date" | "minutes" | "kind" | "note">) =>
    request<Session>("/sessions", { method: "POST", body: JSON.stringify(s) }),
  deleteSession: (id: string) => request<void>(`/sessions/${id}`, { method: "DELETE" }),
};

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

export type Me = {
  id: string;
  email: string;
  emailVerified: boolean;
  username: string;
  displayName: string;
  profilePublic: boolean;
  hasPassword: boolean;
  hasGoogle: boolean;
  createdAt: string;
};

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
    cache: "no-store",
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new ApiError(body?.error ?? `Request failed (${res.status})`, res.status, body?.code);
  }
  return res.status === 204 ? (undefined as T) : res.json();
}

const post = <T>(path: string, body?: unknown, method = "POST") =>
  request<T>(path, { method, body: body === undefined ? undefined : JSON.stringify(body) });

export const auth = {
  providers: () => request<{ google: boolean }>("/auth/providers"),
  signup: (b: { email: string; username: string; password: string }) => post<void>("/auth/signup", b),
  login: (b: { email: string; password: string }) => post<Me>("/auth/login", b),
  logout: () => post<void>("/auth/logout"),
  verify: (token: string) => post<Me>("/auth/verify", { token }),
  resendVerification: (email: string) => post<void>("/auth/resend-verification", { email }),
  forgot: (email: string) => post<void>("/auth/forgot", { email }),
  reset: (token: string, password: string) => post<Me>("/auth/reset", { token, password }),
  me: () => request<Me>("/me"),
  updateMe: (b: Partial<Pick<Me, "username" | "displayName" | "profilePublic">>) => post<Me>("/me", b, "PATCH"),
  changePassword: (current: string, next: string) => post<void>("/me/password", { current, new: next }),
  deleteMe: (b: { password?: string; confirm?: string }) => post<void>("/me", b, "DELETE"),
};

export type PublicProfile = {
  username: string;
  displayName: string;
  joinedAt: string;
  stats: Stats;
};

const user = (username: string) => `/users/${encodeURIComponent(username)}`;

export const profiles = {
  get: (username: string, today: string) => request<PublicProfile>(`${user(username)}?today=${today}`),
  contributions: (username: string, filter: string, today: string) =>
    request<Calendar>(`${user(username)}/contributions?filter=${filter}&today=${today}`),
  years: (username: string, today: string) => request<number[]>(`${user(username)}/years?today=${today}`),
};

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

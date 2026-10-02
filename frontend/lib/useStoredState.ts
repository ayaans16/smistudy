"use client";

import { useCallback, useMemo, useSyncExternalStore } from "react";

const listeners = new Set<() => void>();

function subscribe(cb: () => void) {
  listeners.add(cb);
  window.addEventListener("storage", cb);
  return () => {
    listeners.delete(cb);
    window.removeEventListener("storage", cb);
  };
}

function read(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

/** JSON state persisted in localStorage; renders `fallback` on the server and first paint. */
export function useStoredState<T>(key: string, fallback: T): [T, (value: T) => void] {
  const raw = useSyncExternalStore(subscribe, () => read(key), () => null);
  const value = useMemo<T>(() => {
    if (raw == null) return fallback;
    try {
      return { ...fallback, ...JSON.parse(raw) };
    } catch {
      return fallback;
    }
    // fallback is a constant default; re-parse only when the stored string changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [raw]);

  const set = useCallback(
    (next: T) => {
      try {
        localStorage.setItem(key, JSON.stringify(next));
      } catch {}
      listeners.forEach((l) => l());
    },
    [key],
  );
  return [value, set];
}

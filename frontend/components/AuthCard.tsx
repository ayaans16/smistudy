"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { auth } from "@/lib/api";
import { AngelBuddy, SmiBuddy } from "./Mascots";
import SiteHeader from "./SiteHeader";

/** Centered card used by every sign-in related page. */
export function AuthCard({ title, subtitle, children }: { title: string; subtitle?: string; children: React.ReactNode }) {
  return (
    <div className="mx-auto w-full max-w-6xl px-4 pb-16 sm:px-6">
      <SiteHeader />
      <main className="mx-auto mt-6 w-full max-w-md">
        <div className="relative rounded-3xl border border-line bg-card p-8 shadow-sm">
          <div className="absolute -top-10 left-1/2 flex -translate-x-1/2 items-end gap-1">
            <SmiBuddy className="h-16 w-16" />
            <AngelBuddy className="h-14 w-14" />
          </div>
          <h1 className="mt-6 text-center text-2xl font-extrabold">{title}</h1>
          {subtitle && <p className="mt-1 text-center text-sm text-muted">{subtitle}</p>}
          <div className="mt-6">{children}</div>
        </div>
      </main>
    </div>
  );
}

export function Field({
  label,
  hint,
  ...props
}: { label: string; hint?: string } & React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <label className="block">
      <span className="text-sm font-semibold">{label}</span>
      <input
        {...props}
        className="mt-1 w-full rounded-xl border border-line bg-bg px-3 py-2 outline-none transition focus:border-accent disabled:opacity-60"
      />
      {hint && <span className="mt-1 block text-xs text-muted">{hint}</span>}
    </label>
  );
}

export function SubmitButton({ busy, children }: { busy: boolean; children: React.ReactNode }) {
  return (
    <button
      disabled={busy}
      className="w-full rounded-full bg-accent py-2.5 font-bold text-[#1f2a14] shadow-sm transition hover:brightness-105 disabled:opacity-60"
    >
      {busy ? "One moment…" : children}
    </button>
  );
}

export function FormMessage({ error, success }: { error?: string | null; success?: string | null }) {
  if (error) return <p role="alert" className="rounded-xl bg-red-500/10 px-3 py-2 text-sm text-red-600 dark:text-red-400">{error}</p>;
  if (success) return <p role="status" className="rounded-xl bg-accent-soft px-3 py-2 text-sm text-accent-strong">{success}</p>;
  return null;
}

/** "Continue with Google", shown only when the server has Google sign-in configured. */
export function GoogleButton() {
  const [enabled, setEnabled] = useState(false);
  useEffect(() => {
    auth.providers().then((p) => setEnabled(p.google)).catch(() => {});
  }, []);
  if (!enabled) return null;
  return (
    <>
      <a
        href="/api/auth/google/start"
        className="flex w-full items-center justify-center gap-2 rounded-full border border-line py-2.5 font-bold transition hover:border-muted"
      >
        <svg viewBox="0 0 48 48" className="h-5 w-5" aria-hidden="true">
          <path fill="#FFC107" d="M43.6 20.5H42V20H24v8h11.3C33.7 32.7 29.2 36 24 36c-6.6 0-12-5.4-12-12s5.4-12 12-12c3.1 0 5.8 1.2 7.9 3.1l5.7-5.7C34 6.1 29.3 4 24 4 12.9 4 4 12.9 4 24s8.9 20 20 20 20-8.9 20-20c0-1.3-.1-2.4-.4-3.5z" />
          <path fill="#FF3D00" d="m6.3 14.7 6.6 4.8C14.7 15.1 19 12 24 12c3.1 0 5.8 1.2 7.9 3.1l5.7-5.7C34 6.1 29.3 4 24 4 16.3 4 9.7 8.3 6.3 14.7z" />
          <path fill="#4CAF50" d="M24 44c5.2 0 9.9-2 13.4-5.2l-6.2-5.2C29.2 35.1 26.7 36 24 36c-5.2 0-9.6-3.3-11.3-7.9l-6.5 5C9.5 39.6 16.2 44 24 44z" />
          <path fill="#1976D2" d="M43.6 20.5H42V20H24v8h11.3c-.8 2.2-2.2 4.2-4.1 5.6l6.2 5.2C37 39.2 44 34 44 24c0-1.3-.1-2.4-.4-3.5z" />
        </svg>
        Continue with Google
      </a>
      <p className="mt-2 text-center text-xs text-muted">
        By continuing with Google, you confirm you&apos;re 13 or older and agree to our{" "}
        <Link href="/terms" className="font-semibold text-accent-strong hover:underline">Terms</Link> and{" "}
        <Link href="/privacy" className="font-semibold text-accent-strong hover:underline">Privacy Policy</Link>.
      </p>
      <div className="my-5 flex items-center gap-3 text-xs font-semibold uppercase tracking-widest text-muted">
        <span className="h-px flex-1 bg-line" />
        or
        <span className="h-px flex-1 bg-line" />
      </div>
    </>
  );
}

export function errorMessage(e: unknown) {
  return e instanceof Error ? e.message : "Something went wrong. Please try again.";
}

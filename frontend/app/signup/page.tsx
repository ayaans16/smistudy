"use client";

import Link from "next/link";
import { useState } from "react";
import { AuthCard, Field, FormMessage, GoogleButton, SubmitButton, errorMessage } from "@/components/AuthCard";
import { auth } from "@/lib/api";
import { useMe } from "@/lib/useMe";

export default function SignupPage() {
  useMe({ guestOnly: true });
  const [form, setForm] = useState({ username: "", email: "", password: "" });
  const [agreed, setAgreed] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);
  const [resent, setResent] = useState(false);

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm({ ...form, [k]: e.target.value });

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await auth.signup({ ...form, acceptTerms: agreed });
      setSent(true);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  if (sent) {
    return (
      <AuthCard title="Check your email" subtitle={`We sent a link to ${form.email}. Click it to finish signing up.`}>
        <FormMessage success={resent ? "Sent again. It can take a minute to arrive." : null} />
        <button
          onClick={() => auth.resendVerification(form.email).finally(() => setResent(true))}
          className="mt-3 w-full rounded-full border border-line py-2.5 font-bold transition hover:border-muted"
        >
          Resend email
        </button>
        <p className="mt-4 text-center text-sm text-muted">
          Wrong address? <button onClick={() => setSent(false)} className="font-bold text-accent-strong hover:underline">Go back</button>
        </p>
      </AuthCard>
    );
  }

  return (
    <AuthCard title="Join smistudy" subtitle="Track your study hours and grow your garden.">
      <GoogleButton />
      <form onSubmit={submit} className="space-y-4">
        <Field
          label="Username"
          autoComplete="username"
          required
          minLength={3}
          maxLength={20}
          pattern="[A-Za-z0-9_]+"
          hint="3–20 letters, numbers or _. This is your public profile name."
          value={form.username}
          onChange={set("username")}
        />
        <Field label="Email" type="email" autoComplete="email" required value={form.email} onChange={set("email")} />
        <Field
          label="Password"
          type="password"
          autoComplete="new-password"
          required
          minLength={10}
          maxLength={128}
          hint="At least 10 characters."
          value={form.password}
          onChange={set("password")}
        />
        <label className="flex items-start gap-2 text-sm">
          <input
            type="checkbox"
            required
            checked={agreed}
            onChange={(e) => setAgreed(e.target.checked)}
            className="mt-0.5 h-4 w-4 shrink-0 accent-[var(--accent-strong)]"
          />
          <span>
            I&apos;m 13 or older and I agree to the{" "}
            <Link href="/terms" target="_blank" className="font-bold text-accent-strong hover:underline">Terms of Service</Link> and{" "}
            <Link href="/privacy" target="_blank" className="font-bold text-accent-strong hover:underline">Privacy Policy</Link>.
          </span>
        </label>
        <FormMessage error={error} />
        <SubmitButton busy={busy}>Create account</SubmitButton>
      </form>
      <p className="mt-5 text-center text-sm text-muted">
        Already have an account?{" "}
        <Link href="/login" className="font-bold text-accent-strong hover:underline">
          Log in
        </Link>
      </p>
    </AuthCard>
  );
}

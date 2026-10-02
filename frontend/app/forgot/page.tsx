"use client";

import Link from "next/link";
import { useState } from "react";
import { AuthCard, Field, FormMessage, SubmitButton, errorMessage } from "@/components/AuthCard";
import { auth } from "@/lib/api";

export default function ForgotPage() {
  const [email, setEmail] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await auth.forgot(email);
      setSent(true);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <AuthCard title="Forgot your password?" subtitle="We'll email you a link to pick a new one.">
      {sent ? (
        <FormMessage success={`If ${email} has an account, a reset link is on its way. It expires in 1 hour.`} />
      ) : (
        <form onSubmit={submit} className="space-y-4">
          <Field label="Email" type="email" autoComplete="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
          <FormMessage error={error} />
          <SubmitButton busy={busy}>Send reset link</SubmitButton>
        </form>
      )}
      <p className="mt-5 text-center text-sm">
        <Link href="/login" className="font-bold text-accent-strong hover:underline">
          Back to log in
        </Link>
      </p>
    </AuthCard>
  );
}

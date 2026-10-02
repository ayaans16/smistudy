"use client";

import { useRouter } from "next/navigation";
import { use, useState } from "react";
import { AuthCard, Field, FormMessage, SubmitButton, errorMessage } from "@/components/AuthCard";
import { auth } from "@/lib/api";

export default function ResetPage({ searchParams }: PageProps<"/reset">) {
  const { token } = use(searchParams);
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (password !== confirm) return setError("Those passwords don't match.");
    if (typeof token !== "string") return setError("This link is missing its token. Request a new one.");
    setBusy(true);
    setError(null);
    try {
      await auth.reset(token, password);
      router.replace("/");
    } catch (err) {
      setError(errorMessage(err));
      setBusy(false);
    }
  }

  return (
    <AuthCard title="Choose a new password" subtitle="This signs you out on your other devices.">
      <form onSubmit={submit} className="space-y-4">
        <Field
          label="New password"
          type="password"
          autoComplete="new-password"
          required
          minLength={10}
          maxLength={128}
          hint="At least 10 characters."
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <Field
          label="Confirm password"
          type="password"
          autoComplete="new-password"
          required
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
        />
        <FormMessage error={error} />
        <SubmitButton busy={busy}>Save password</SubmitButton>
      </form>
    </AuthCard>
  );
}

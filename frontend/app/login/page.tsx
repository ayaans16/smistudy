"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { use, useState } from "react";
import { AuthCard, Field, FormMessage, GoogleButton, SubmitButton, errorMessage } from "@/components/AuthCard";
import { auth } from "@/lib/api";
import { useMe } from "@/lib/useMe";

export default function LoginPage({ searchParams }: PageProps<"/login">) {
  const params = use(searchParams);
  useMe({ guestOnly: true });
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(
    params.error === "google" ? "Google sign-in didn't work. Please try again." : null,
  );

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await auth.login({ email, password });
      router.replace("/");
    } catch (err) {
      setError(errorMessage(err));
      setBusy(false);
    }
  }

  return (
    <AuthCard title="Welcome back" subtitle="Log in to keep your streak going.">
      <GoogleButton />
      <form onSubmit={submit} className="space-y-4">
        <Field label="Email" type="email" autoComplete="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
        <Field
          label="Password"
          type="password"
          autoComplete="current-password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <FormMessage error={error} />
        <SubmitButton busy={busy}>Log in</SubmitButton>
      </form>
      <div className="mt-5 flex justify-between text-sm">
        <Link href="/forgot" className="text-muted hover:text-fg">
          Forgot password?
        </Link>
        <Link href="/signup" className="font-bold text-accent-strong hover:underline">
          Create an account
        </Link>
      </div>
    </AuthCard>
  );
}

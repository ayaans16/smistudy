"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { use, useEffect, useRef, useState } from "react";
import { AuthCard, FormMessage, errorMessage } from "@/components/AuthCard";
import { auth } from "@/lib/api";

export default function VerifyPage({ searchParams }: PageProps<"/verify">) {
  const { token } = use(searchParams);
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const started = useRef(false);

  useEffect(() => {
    // Tokens are single-use, so guard against React running this effect twice in dev.
    if (started.current) return;
    started.current = true;
    if (typeof token !== "string") {
      Promise.resolve().then(() => setError("This link is missing its token."));
      return;
    }
    auth
      .verify(token)
      .then(() => router.replace("/"))
      .catch((e) => setError(errorMessage(e)));
  }, [token, router]);

  return (
    <AuthCard title={error ? "Hmm, that didn't work" : "Confirming your email…"}>
      <FormMessage error={error} />
      {error && (
        <p className="mt-4 text-center text-sm text-muted">
          Try <Link href="/login" className="font-bold text-accent-strong hover:underline">logging in</Link>. We&apos;ll send a fresh link if you still need one.
        </p>
      )}
    </AuthCard>
  );
}

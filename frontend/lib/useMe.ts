"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { ApiError, auth, type Me } from "./api";

/**
 * Loads the signed-in user. With `required`, anonymous visitors are sent to /login;
 * with `guestOnly`, signed-in users are sent home (for the login/signup pages).
 */
export function useMe({ required = false, guestOnly = false } = {}) {
  const router = useRouter();
  const [me, setMe] = useState<Me | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    auth
      .me()
      .then((u) => {
        if (cancelled) return;
        if (guestOnly) router.replace("/");
        else setMe(u);
      })
      .catch((e) => {
        if (cancelled) return;
        if (e instanceof ApiError && e.status === 401 && required) router.replace("/login");
      })
      .finally(() => !cancelled && setLoading(false));
    return () => {
      cancelled = true;
    };
  }, [required, guestOnly, router]);

  return { me, setMe, loading };
}

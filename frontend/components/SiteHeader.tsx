"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { auth, type Me } from "@/lib/api";
import { SmiBuddy } from "./Mascots";
import ThemeToggle from "./ThemeToggle";

export default function SiteHeader({ me }: { me?: Me | null }) {
  return (
    <header className="flex items-center justify-between py-6">
      <Link href="/" className="flex items-center gap-2">
        <SmiBuddy className="h-10 w-10" />
        <span className="text-2xl font-extrabold tracking-tight">
          smi<span className="text-accent-strong">study</span>
        </span>
      </Link>
      <div className="flex items-center gap-2">
        <ThemeToggle />
        {me && <UserMenu me={me} />}
      </div>
    </header>
  );
}

function UserMenu({ me }: { me: Me }) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const close = (e: MouseEvent) => !ref.current?.contains(e.target as Node) && setOpen(false);
    document.addEventListener("mousedown", close);
    return () => document.removeEventListener("mousedown", close);
  }, [open]);

  async function logout() {
    await auth.logout().catch(() => {});
    router.replace("/login");
  }

  const initial = (me.displayName || me.username).charAt(0).toUpperCase();
  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen((o) => !o)}
        aria-label="Account menu"
        aria-expanded={open}
        className="flex h-10 w-10 items-center justify-center rounded-full border border-line bg-accent-soft font-extrabold text-accent-strong transition hover:border-accent"
      >
        {initial}
      </button>
      {open && (
        <div className="absolute right-0 z-20 mt-2 w-52 overflow-hidden rounded-2xl border border-line bg-card py-1 text-sm shadow-lg">
          <div className="border-b border-line px-4 py-2">
            <div className="truncate font-bold">{me.displayName || me.username}</div>
            <div className="truncate text-muted">@{me.username}</div>
          </div>
          <Link href="/settings" className="block px-4 py-2 hover:bg-accent-soft" onClick={() => setOpen(false)}>
            Settings
          </Link>
          <button onClick={logout} className="block w-full px-4 py-2 text-left hover:bg-accent-soft">
            Log out
          </button>
        </div>
      )}
    </div>
  );
}

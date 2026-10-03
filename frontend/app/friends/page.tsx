"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { use, useCallback, useEffect, useState, useSyncExternalStore } from "react";
import { errorMessage } from "@/components/AuthCard";
import { PersonLink } from "@/components/FollowControls";
import { AngelBuddy } from "@/components/Mascots";
import SiteHeader from "@/components/SiteHeader";
import { social, type Follower, type FollowUser, type FriendStats } from "@/lib/api";
import { formatHours, formatMinutes, localDateKey } from "@/lib/format";
import { useMe } from "@/lib/useMe";

type Tab = "following" | "followers" | "blocked";
const TABS: { id: Tab; label: string }[] = [
  { id: "following", label: "Following" },
  { id: "followers", label: "Followers" },
  { id: "blocked", label: "Blocked" },
];
const noop = () => () => {};

export default function FriendsPage({ searchParams }: PageProps<"/friends">) {
  const params = use(searchParams);
  const initial = TABS.some((t) => t.id === params.tab) ? (params.tab as Tab) : "following";
  const router = useRouter();
  const { me } = useMe({ required: true });
  const today = useSyncExternalStore(noop, () => localDateKey(), () => "");
  const [tab, setTab] = useState<Tab>(initial);

  function choose(t: Tab) {
    setTab(t);
    router.replace(`/friends${t === "following" ? "" : `?tab=${t}`}`, { scroll: false });
  }

  return (
    <div className="mx-auto w-full max-w-6xl px-4 pb-16 sm:px-6">
      <SiteHeader me={me} />
      {me && today && (
        <main className="mx-auto flex max-w-2xl flex-col gap-6">
          <h1 className="text-2xl font-extrabold">Friends</h1>

          {!me.profilePublic && (
            <div className="rounded-2xl border border-blush bg-blush-soft p-4 text-sm">
              Your profile is private, so you can&apos;t follow people and won&apos;t show up in anyone&apos;s lists.{" "}
              <Link href="/settings" className="font-bold underline">
                Make it public in Settings
              </Link>
            </div>
          )}

          <div className="flex gap-1 rounded-full bg-card p-1 text-sm font-semibold shadow-sm" role="tablist">
            {TABS.map((t) => (
              <button
                key={t.id}
                role="tab"
                aria-selected={tab === t.id}
                onClick={() => choose(t.id)}
                className={`flex-1 rounded-full px-3 py-1.5 transition ${tab === t.id ? "bg-accent text-[#1f2a14]" : "text-muted hover:text-fg"}`}
              >
                {t.label}
              </button>
            ))}
          </div>

          {tab === "following" && <FollowingTab today={today} username={me.username} />}
          {tab === "followers" && <FollowersTab />}
          {tab === "blocked" && <BlockedTab />}
        </main>
      )}
    </div>
  );
}

/** Loads a list and gives back a reload function for after actions. */
function useList<T>(load: () => Promise<T[]>) {
  const [items, setItems] = useState<T[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const reload = useCallback(() => {
    load()
      .then((list) => {
        setItems(list);
        setError(null);
      })
      .catch((e) => setError(errorMessage(e)));
  }, [load]);
  useEffect(reload, [reload]);
  const act = (action: () => Promise<unknown>) => action().then(reload, (e) => setError(errorMessage(e)));
  return { items, error, act };
}

function Card({ children }: { children: React.ReactNode }) {
  return <section className="rounded-3xl border border-line bg-card p-4 shadow-sm sm:p-6">{children}</section>;
}

function Empty({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-3 p-2 text-sm text-muted">
      <AngelBuddy className="h-12 w-12 shrink-0" />
      <span>{children}</span>
    </div>
  );
}

const smallButton = "shrink-0 rounded-full border border-line px-3 py-1 text-xs font-bold transition hover:border-muted";

function FollowingTab({ today, username }: { today: string; username: string }) {
  const load = useCallback(() => social.myFollowing(today), [today]);
  const { items, error, act } = useList<FriendStats>(load);

  return (
    <Card>
      <div className="flex items-baseline justify-between px-2">
        <h2 className="font-bold">This week&apos;s leaderboard</h2>
        <span className="text-xs text-muted">Ranked by study time since Sunday</span>
      </div>
      {error && <p className="mt-2 px-2 text-sm text-red-500">{error}</p>}
      {items?.length === 0 && (
        <div className="mt-3">
          <Empty>
            You&apos;re not following anyone yet. Ask friends for their profile link, or share yours:{" "}
            <Link href={`/u/${username}`} className="font-bold text-accent-strong hover:underline">
              smistudy.ca/u/{username}
            </Link>
          </Empty>
        </div>
      )}
      <ol className="mt-3 space-y-1">
        {items?.map((f, i) => (
          <li key={f.username} className="flex items-center gap-2">
            <span className={`w-7 shrink-0 text-center text-sm font-extrabold ${i < 3 && f.weekMinutes > 0 ? "text-accent-strong" : "text-muted"}`}>
              {i + 1}
            </span>
            <div className="min-w-0 flex-1">
              <PersonLink person={f} />
            </div>
            <div className="shrink-0 text-right text-xs">
              <div className="font-bold tabular-nums">{formatMinutes(f.weekMinutes)}</div>
              <div className="text-muted">
                {f.currentStreak > 0 ? `🔥 ${f.currentStreak}d` : "—"} · {formatHours(f.totalMinutes)}h total
              </div>
            </div>
            <button onClick={() => act(() => social.unfollow(f.username))} className={`${smallButton} hover:text-red-500`}>
              Unfollow
            </button>
          </li>
        ))}
      </ol>
    </Card>
  );
}

function FollowersTab() {
  const { items, error, act } = useList<Follower>(social.myFollowers);

  function block(f: FollowUser) {
    if (window.confirm(`Block @${f.username}? Neither of you will be able to follow the other.`)) act(() => social.block(f.username));
  }

  return (
    <Card>
      {error && <p className="mb-2 px-2 text-sm text-red-500">{error}</p>}
      {items?.length === 0 && <Empty>No followers yet. When someone follows you, they&apos;ll show up here.</Empty>}
      <ul className="space-y-1">
        {items?.map((f) => (
          <li key={f.username} className="flex flex-wrap items-center gap-2">
            <div className="min-w-0 flex-1">
              <PersonLink person={f} />
            </div>
            {!f.followingBack && (
              <button onClick={() => act(() => social.follow(f.username))} className="shrink-0 rounded-full bg-accent px-3 py-1 text-xs font-bold text-[#1f2a14]">
                Follow back
              </button>
            )}
            <button onClick={() => act(() => social.removeFollower(f.username))} className={smallButton}>
              Remove
            </button>
            <button onClick={() => block(f)} className={`${smallButton} hover:text-red-500`}>
              Block
            </button>
          </li>
        ))}
      </ul>
    </Card>
  );
}

function BlockedTab() {
  const { items, error, act } = useList<FollowUser>(social.blocked);
  return (
    <Card>
      {error && <p className="mb-2 px-2 text-sm text-red-500">{error}</p>}
      <p className="px-2 text-xs text-muted">Blocked people can&apos;t follow you, and you won&apos;t appear in each other&apos;s lists. They aren&apos;t told.</p>
      {items?.length === 0 && <p className="mt-3 px-2 text-sm text-muted">You haven&apos;t blocked anyone.</p>}
      <ul className="mt-2 space-y-1">
        {items?.map((b) => (
          <li key={b.username} className="flex items-center gap-2">
            <div className="min-w-0 flex-1">
              <PersonLink person={b} />
            </div>
            <button onClick={() => act(() => social.unblock(b.username))} className={smallButton}>
              Unblock
            </button>
          </li>
        ))}
      </ul>
    </Card>
  );
}

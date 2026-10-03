"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ApiError, social, type FollowUser, type PublicProfile } from "@/lib/api";
import { CONTACT_EMAIL } from "@/lib/legal";
import { errorMessage } from "./AuthCard";

/** Follow / Following button for someone else's profile. */
export function FollowButton({ profile, signedIn, onChange }: { profile: PublicProfile; signedIn: boolean; onChange: () => void }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<{ text: string; needsPublic?: boolean } | null>(null);

  if (!signedIn) {
    return (
      <Link href="/login" className="rounded-full bg-accent px-5 py-2 text-sm font-bold text-[#1f2a14] transition hover:brightness-105">
        Log in to follow
      </Link>
    );
  }
  if (!profile.viewer || profile.viewer.isSelf) {
    return (
      <Link href="/settings" className="rounded-full border border-line px-5 py-2 text-sm font-bold transition hover:border-accent">
        Edit profile
      </Link>
    );
  }
  const following = profile.viewer.following;

  async function toggle() {
    setBusy(true);
    setError(null);
    try {
      await (following ? social.unfollow(profile.username) : social.follow(profile.username));
      onChange();
    } catch (e) {
      setError({ text: errorMessage(e), needsPublic: e instanceof ApiError && e.code === "profile_private" });
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col items-start gap-1">
      <div className="flex items-center gap-2">
        {profile.viewer.followsYou && <span className="rounded-full bg-bg px-2 py-0.5 text-xs font-semibold text-muted">Follows you</span>}
        <button
          onClick={toggle}
          disabled={busy}
          className={`group rounded-full px-5 py-2 text-sm font-bold transition disabled:opacity-60 ${
            following ? "border border-line hover:border-red-400 hover:text-red-500" : "bg-accent text-[#1f2a14] hover:brightness-105"
          }`}
        >
          {following ? (
            <>
              <span className="group-hover:hidden">Following</span>
              <span className="hidden group-hover:inline">Unfollow</span>
            </>
          ) : profile.viewer.followsYou ? (
            "Follow back"
          ) : (
            "Follow"
          )}
        </button>
      </div>
      {error && (
        <p className="max-w-xs text-xs text-red-500">
          {error.text}
          {error.needsPublic && (
            <>
              {" "}
              <Link href="/settings" className="font-bold underline">
                Open Settings
              </Link>
            </>
          )}
        </p>
      )}
    </div>
  );
}

/** "12 followers · 5 following": each opens that list. */
export function FollowCounts({
  profile,
  open,
  onOpen,
}: {
  profile: PublicProfile;
  open: "followers" | "following" | null;
  onOpen: (list: "followers" | "following" | null) => void;
}) {
  const item = (list: "followers" | "following", count: number, label: string) => (
    <button
      onClick={() => onOpen(open === list ? null : list)}
      aria-expanded={open === list}
      className={`transition hover:text-accent-strong ${open === list ? "text-accent-strong" : ""}`}
    >
      <strong className="text-fg">{count}</strong> {label}
    </button>
  );
  return (
    <p className="mt-2 flex gap-3 text-sm text-muted">
      {item("followers", profile.followers, profile.followers === 1 ? "follower" : "followers")}
      <span aria-hidden="true">·</span>
      {item("following", profile.following, "following")}
    </p>
  );
}

/** The followers or following list of a public profile. */
export function FollowListPanel({ username, list, version }: { username: string; list: "followers" | "following"; version: number }) {
  const [people, setPeople] = useState<FollowUser[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    (list === "followers" ? social.followers(username) : social.following(username))
      .then((p) => !cancelled && setPeople(p))
      .catch(() => !cancelled && setPeople([]));
    return () => {
      cancelled = true;
    };
  }, [username, list, version]);

  return (
    <section className="rounded-3xl border border-line bg-card p-6 shadow-sm">
      <h2 className="text-lg font-bold capitalize">{list}</h2>
      {people && people.length === 0 && (
        <p className="mt-3 text-sm text-muted">{list === "followers" ? "No followers yet." : "Not following anyone yet."}</p>
      )}
      <ul className="mt-3 grid gap-2 sm:grid-cols-2">
        {people?.map((p) => (
          <li key={p.username}>
            <PersonLink person={p} />
          </li>
        ))}
      </ul>
    </section>
  );
}

export function PersonLink({ person }: { person: FollowUser }) {
  const initial = (person.displayName || person.username).charAt(0).toUpperCase();
  return (
    <Link href={`/u/${person.username}`} className="flex items-center gap-3 rounded-2xl p-2 transition hover:bg-bg">
      <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-accent-soft font-extrabold text-accent-strong">
        {initial}
      </span>
      <span className="min-w-0">
        <span className="block truncate text-sm font-bold">{person.displayName || person.username}</span>
        <span className="block truncate text-xs text-muted">@{person.username}</span>
      </span>
    </Link>
  );
}

/** Block and report links for someone else's profile. */
export function ProfileSafetyActions({ username, onBlocked }: { username: string; onBlocked: () => void }) {
  const [error, setError] = useState<string | null>(null);

  async function block() {
    if (!window.confirm(`Block @${username}? They'll be removed from your followers and following, and neither of you can follow the other.`)) return;
    try {
      await social.block(username);
      onBlocked();
    } catch (e) {
      setError(errorMessage(e));
    }
  }

  const report = `mailto:${CONTACT_EMAIL}?subject=${encodeURIComponent(`Report @${username}`)}&body=${encodeURIComponent(
    `Profile: https://smistudy.ca/u/${username}\n\nWhat's wrong:\n`,
  )}`;

  return (
    <div className="flex justify-end gap-4 text-xs text-muted">
      {error && <span className="text-red-500">{error}</span>}
      <button onClick={block} className="hover:text-red-500">
        Block
      </button>
      <a href={report} className="hover:text-fg">
        Report
      </a>
    </div>
  );
}

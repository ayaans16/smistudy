"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { use, useEffect, useState, useSyncExternalStore } from "react";
import { FollowButton, FollowCounts, FollowListPanel, ProfileSafetyActions } from "@/components/FollowControls";
import { AngelBuddy, SmiBuddy } from "@/components/Mascots";
import SiteHeader from "@/components/SiteHeader";
import StudyGraph from "@/components/StudyGraph";
import { ApiError, profiles, type PublicProfile } from "@/lib/api";
import { formatHours, formatMinutes, localDateKey } from "@/lib/format";
import { useMe } from "@/lib/useMe";

const noop = () => () => {};

export default function ProfilePage({ params }: PageProps<"/u/[username]">) {
  const { username } = use(params);
  const { me } = useMe();
  const router = useRouter();
  const today = useSyncExternalStore(noop, () => localDateKey(), () => "");
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [missing, setMissing] = useState(false);
  const [openList, setOpenList] = useState<"followers" | "following" | null>(null);
  const [version, setVersion] = useState(0); // bump to reload after following/unfollowing
  const reload = () => setVersion((v) => v + 1);

  useEffect(() => {
    if (!today) return;
    let cancelled = false;
    profiles
      .get(username, today)
      .then((p) => !cancelled && setProfile(p))
      .catch((e) => !cancelled && e instanceof ApiError && e.status === 404 && setMissing(true));
    return () => {
      cancelled = true;
    };
  }, [username, today, version]);

  useEffect(() => {
    if (profile) document.title = `${profile.displayName || profile.username} — smistudy`;
  }, [profile]);

  return (
    <div className="mx-auto w-full max-w-6xl px-4 pb-16 sm:px-6">
      <SiteHeader me={me} />

      {missing && (
        <main className="mx-auto mt-10 max-w-md rounded-3xl border border-line bg-card p-8 text-center shadow-sm">
          <AngelBuddy className="mx-auto h-20 w-20" />
          <h1 className="mt-4 text-xl font-extrabold">This profile isn&apos;t public</h1>
          <p className="mt-1 text-sm text-muted">It may be private, or the username may not exist.</p>
          <Link href={me ? "/" : "/signup"} className="mt-5 inline-block font-bold text-accent-strong hover:underline">
            {me ? "Back to studying" : "Start tracking your own study hours"}
          </Link>
        </main>
      )}

      {profile && (
        <main className="flex flex-col gap-8">
          <section className="flex flex-col gap-6 rounded-3xl border border-line bg-card p-6 shadow-sm sm:flex-row sm:items-center">
            <div className="flex items-center gap-4">
              <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-full bg-accent-soft">
                <SmiBuddy className="h-14 w-14" />
              </div>
              <div>
                <h1 className="text-2xl font-extrabold">{profile.displayName || profile.username}</h1>
                <p className="text-muted">@{profile.username}</p>
                <p className="mt-1 text-xs text-muted">
                  Studying since{" "}
                  {new Date(profile.joinedAt).toLocaleDateString(undefined, { month: "long", year: "numeric" })}
                </p>
                <FollowCounts profile={profile} open={openList} onOpen={setOpenList} />
                <div className="mt-3">
                  <FollowButton profile={profile} signedIn={Boolean(me)} onChange={reload} />
                </div>
              </div>
            </div>
            <div className="grid flex-1 grid-cols-2 gap-3 sm:grid-cols-4">
              {[
                ["All time", `${formatHours(profile.stats.totalMinutes)}h`],
                ["This week", formatMinutes(profile.stats.weekMinutes)],
                ["Current streak", `${profile.stats.currentStreak}d`],
                ["Longest streak", `${profile.stats.longestStreak}d`],
              ].map(([label, value]) => (
                <div key={label} className="rounded-2xl bg-bg px-4 py-3">
                  <div className="text-xs font-semibold uppercase tracking-wide text-muted">{label}</div>
                  <div className="mt-1 text-xl font-extrabold tabular-nums">{value}</div>
                </div>
              ))}
            </div>
          </section>
          {openList && <FollowListPanel username={profile.username} list={openList} version={version} />}
          <StudyGraph today={today} username={profile.username} />
          {me && profile.viewer && !profile.viewer.isSelf && (
            <ProfileSafetyActions username={profile.username} onBlocked={() => router.push("/friends?tab=blocked")} />
          )}
        </main>
      )}
    </div>
  );
}

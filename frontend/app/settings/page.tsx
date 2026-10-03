"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Field, FormMessage, errorMessage } from "@/components/AuthCard";
import SiteHeader from "@/components/SiteHeader";
import { auth, type Me } from "@/lib/api";
import { useMe } from "@/lib/useMe";

export default function SettingsPage() {
  const { me, setMe } = useMe({ required: true });

  return (
    <div className="mx-auto w-full max-w-6xl px-4 pb-16 sm:px-6">
      <SiteHeader me={me} />
      {me && (
        <main className="mx-auto flex max-w-xl flex-col gap-6">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-extrabold">Settings</h1>
            <Link href="/" className="text-sm font-semibold text-muted hover:text-fg">
              ← Back to studying
            </Link>
          </div>
          <ProfileSection me={me} onSaved={setMe} />
          <VisibilitySection me={me} onSaved={setMe} />
          <PasswordSection me={me} />
          <Section
            title="Your data"
            description="Download a copy of everything smistudy stores about you (your account details and every study session) as a JSON file."
          >
            <a
              href="/api/me/export"
              download
              className="inline-block rounded-full bg-fg px-5 py-2 text-sm font-bold text-bg transition hover:opacity-90"
            >
              Download my data
            </a>
            <p className="mt-3 text-xs text-muted">
              See the <Link href="/privacy" className="font-semibold text-accent-strong hover:underline">Privacy Policy</Link> for
              how we handle your information.
            </p>
          </Section>
          <DeleteSection me={me} />
        </main>
      )}
    </div>
  );
}

function Section({ title, description, children }: { title: string; description?: string; children: React.ReactNode }) {
  return (
    <section className="rounded-3xl border border-line bg-card p-6 shadow-sm">
      <h2 className="text-lg font-bold">{title}</h2>
      {description && <p className="mt-1 text-sm text-muted">{description}</p>}
      <div className="mt-4">{children}</div>
    </section>
  );
}

function SaveButton({ busy, children, danger }: { busy: boolean; children: React.ReactNode; danger?: boolean }) {
  return (
    <button
      disabled={busy}
      className={`rounded-full px-5 py-2 text-sm font-bold transition disabled:opacity-60 ${
        danger ? "bg-red-500 text-white hover:bg-red-600" : "bg-fg text-bg hover:opacity-90"
      }`}
    >
      {busy ? "Saving…" : children}
    </button>
  );
}

function ProfileSection({ me, onSaved }: { me: Me; onSaved: (me: Me) => void }) {
  const [displayName, setDisplayName] = useState(me.displayName);
  const [username, setUsername] = useState(me.username);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ error?: string; success?: string }>({});

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setMsg({});
    try {
      onSaved(await auth.updateMe({ displayName, username }));
      setMsg({ success: "Saved." });
    } catch (err) {
      setMsg({ error: errorMessage(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <Section title="Profile">
      <form onSubmit={save} className="space-y-4">
        <Field label="Display name" maxLength={50} value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
        <Field
          label="Username"
          required
          minLength={3}
          maxLength={20}
          pattern="[A-Za-z0-9_]+"
          hint="3–20 letters, numbers or _."
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
        <Field label="Email" value={me.email} disabled readOnly />
        <FormMessage error={msg.error} success={msg.success} />
        <SaveButton busy={busy}>Save profile</SaveButton>
      </form>
    </Section>
  );
}

function VisibilitySection({ me, onSaved }: { me: Me; onSaved: (me: Me) => void }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const url = typeof window === "undefined" ? `/u/${me.username}` : `${window.location.origin}/u/${me.username}`;

  async function toggle() {
    setBusy(true);
    setError(null);
    try {
      onSaved(await auth.updateMe({ profilePublic: !me.profilePublic }));
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <Section
      title="Public profile"
      description="Show your study graph, streaks and total hours on a page anyone can visit. Your email, notes and individual sessions always stay private."
    >
      <div className="flex items-center justify-between gap-4">
        <span className="text-sm font-semibold">{me.profilePublic ? "Your profile is public" : "Your profile is private"}</span>
        <button
          role="switch"
          aria-checked={me.profilePublic}
          aria-label="Public profile"
          disabled={busy}
          onClick={toggle}
          className={`relative h-7 w-12 shrink-0 rounded-full transition disabled:opacity-60 ${me.profilePublic ? "bg-accent" : "bg-line"}`}
        >
          <span
            className={`absolute top-1 h-5 w-5 rounded-full bg-white shadow transition-all ${me.profilePublic ? "left-6" : "left-1"}`}
          />
        </button>
      </div>
      {me.profilePublic && (
        <div className="mt-4 flex items-center gap-2 rounded-xl bg-bg px-3 py-2 text-sm">
          <Link href={`/u/${me.username}`} className="min-w-0 flex-1 truncate font-semibold text-accent-strong hover:underline">
            {url}
          </Link>
          <button
            onClick={() => navigator.clipboard.writeText(url).then(() => setCopied(true))}
            className="shrink-0 rounded-full border border-line px-3 py-1 text-xs font-bold transition hover:border-muted"
          >
            {copied ? "Copied!" : "Copy link"}
          </button>
        </div>
      )}
      {me.profilePublic && <CardEmbed username={me.username} />}
      {error && <div className="mt-3"><FormMessage error={error} /></div>}
    </Section>
  );
}

/** Copyable Markdown for the SVG stats card, e.g. for a GitHub profile README. */
function CardEmbed({ username }: { username: string }) {
  const [theme, setTheme] = useState<"light" | "dark">("light");
  const [copied, setCopied] = useState(false);
  const origin = typeof window === "undefined" ? "https://smistudy.ca" : window.location.origin;
  const card = `/api/users/${username}/card.svg${theme === "dark" ? "?theme=dark" : ""}`;
  const markdown = `[![smistudy stats](${origin}${card})](${origin}/u/${username})`;

  return (
    <div className="mt-5 border-t border-line pt-5">
      <div className="flex items-center justify-between gap-2">
        <h3 className="text-sm font-bold">Stats card</h3>
        <div className="flex gap-1 rounded-full bg-bg p-1 text-xs font-semibold">
          {(["light", "dark"] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTheme(t)}
              className={`rounded-full px-3 py-1 capitalize transition ${theme === t ? "bg-card shadow-sm" : "text-muted"}`}
            >
              {t}
            </button>
          ))}
        </div>
      </div>
      <p className="mt-1 text-xs text-muted">Show your study stats anywhere that supports images, like your GitHub profile README.</p>
      {/* eslint-disable-next-line @next/next/no-img-element -- a live SVG from our own API, not a static asset */}
      <img src={card} alt="Your smistudy stats card" width={495} height={195} className="mt-3 h-auto w-full max-w-[495px]" />
      <div className="mt-3 flex items-center gap-2 rounded-xl bg-bg px-3 py-2 text-xs">
        <code className="min-w-0 flex-1 truncate">{markdown}</code>
        <button
          onClick={() => navigator.clipboard.writeText(markdown).then(() => setCopied(true))}
          className="shrink-0 rounded-full border border-line px-3 py-1 font-bold transition hover:border-muted"
        >
          {copied ? "Copied!" : "Copy Markdown"}
        </button>
      </div>
    </div>
  );
}

function PasswordSection({ me }: { me: Me }) {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<{ error?: string; success?: string }>({});

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setMsg({});
    try {
      await auth.changePassword(current, next);
      setCurrent("");
      setNext("");
      setMsg({ success: "Password updated. Your other devices have been signed out." });
    } catch (err) {
      setMsg({ error: errorMessage(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <Section
      title={me.hasPassword ? "Change password" : "Set a password"}
      description={me.hasPassword ? undefined : "You sign in with Google. Add a password to also log in with your email."}
    >
      <form onSubmit={save} className="space-y-4">
        {me.hasPassword && (
          <Field
            label="Current password"
            type="password"
            autoComplete="current-password"
            required
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
          />
        )}
        <Field
          label="New password"
          type="password"
          autoComplete="new-password"
          required
          minLength={10}
          maxLength={128}
          hint="At least 10 characters."
          value={next}
          onChange={(e) => setNext(e.target.value)}
        />
        <FormMessage error={msg.error} success={msg.success} />
        <SaveButton busy={busy}>{me.hasPassword ? "Change password" : "Set password"}</SaveButton>
      </form>
    </Section>
  );
}

function DeleteSection({ me }: { me: Me }) {
  const router = useRouter();
  const [value, setValue] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function remove(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await auth.deleteMe(me.hasPassword ? { password: value } : { confirm: value });
      router.replace("/signup");
    } catch (err) {
      setError(errorMessage(err));
      setBusy(false);
    }
  }

  return (
    <Section title="Delete account" description="Permanently deletes your account and every study session. This can't be undone.">
      <form onSubmit={remove} className="space-y-4">
        {me.hasPassword ? (
          <Field label="Password" type="password" autoComplete="current-password" required value={value} onChange={(e) => setValue(e.target.value)} />
        ) : (
          <Field label={`Type ${me.username} to confirm`} required value={value} onChange={(e) => setValue(e.target.value)} />
        )}
        <FormMessage error={error} />
        <SaveButton busy={busy} danger>
          Delete my account
        </SaveButton>
      </form>
    </Section>
  );
}

import type { Metadata } from "next";
import Link from "next/link";
import LegalPage from "@/components/LegalPage";
import { CONTACT_EMAIL } from "@/lib/legal";

export const metadata: Metadata = {
  title: "Cookies & Storage — smistudy",
  description: "The cookies and browser storage smistudy uses, and why.",
};

export default function CookiesPage() {
  return (
    <LegalPage
      title="Cookies & Storage"
      summary={
        <p>
          <strong>The short version:</strong> smistudy only uses the cookies it needs to keep you signed in and secure, plus
          two settings saved in your own browser. No analytics, no ads, no tracking, so there&apos;s nothing to opt out of
          and no cookie banner.
        </p>
      }
    >
      <p>
        Cookies are small files a website saves in your browser. &quot;Local storage&quot; is a similar feature that keeps data
        on your device. This notice lists everything smistudy stores in your browser and why. It complements our{" "}
        <Link href="/privacy">Privacy Policy</Link>.
      </p>

      <h2>1. Strictly necessary cookies</h2>
      <p>These are required for the site to work. Without them you couldn&apos;t sign in.</p>
      <table>
        <thead>
          <tr><th>Name</th><th>Purpose</th><th>Lasts</th><th>Set by</th></tr>
        </thead>
        <tbody>
          <tr>
            <td><code>__Host-smistudy_session</code></td>
            <td>Keeps you signed in. Contains a random token, not your personal details.</td>
            <td>30 days, renewed while you use the site; removed when you log out</td>
            <td>smistudy</td>
          </tr>
          <tr>
            <td><code>smistudy_oauth</code></td>
            <td>Security check while you sign in with Google (prevents forged sign-in requests).</td>
            <td>10 minutes, only during Google sign-in</td>
            <td>smistudy</td>
          </tr>
          <tr>
            <td><code>__cf_bm</code>, <code>_cfuvid</code>, <code>cf_clearance</code></td>
            <td>Set only when needed by Cloudflare, our security provider, to tell people from bots and block attacks.</td>
            <td>From the browsing session up to about a year, depending on the cookie</td>
            <td>Cloudflare</td>
          </tr>
        </tbody>
      </table>

      <h2>2. Preferences saved on your device</h2>
      <p>These are saved in your browser&apos;s local storage and are never sent to our server.</p>
      <table>
        <thead>
          <tr><th>Name</th><th>Purpose</th><th>Lasts</th></tr>
        </thead>
        <tbody>
          <tr><td><code>smistudy-theme</code></td><td>Remembers light or dark mode.</td><td>Until you clear your browser data</td></tr>
          <tr><td><code>smistudy-pomodoro</code></td><td>Remembers your timer and break lengths.</td><td>Until you clear your browser data</td></tr>
        </tbody>
      </table>

      <h2>3. What we don&apos;t use</h2>
      <p>
        smistudy does not use analytics or advertising cookies, social-media pixels, fingerprinting or any other technology
        that tracks you across sites or builds a profile of you. If that ever changes, it would be off by default, and we would
        ask for your consent first and update this notice.
      </p>

      <h2>4. Your choices</h2>
      <p>
        You can view, block or delete cookies and local storage in your browser&apos;s settings. Blocking the strictly necessary
        cookies will stop you from signing in. Clearing local storage just resets your theme and timer preferences.
      </p>

      <h2>5. Contact</h2>
      <p>
        Questions: <a href={`mailto:${CONTACT_EMAIL}`}>{CONTACT_EMAIL}</a>.
      </p>
    </LegalPage>
  );
}

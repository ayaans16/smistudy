import type { Metadata } from "next";
import Link from "next/link";
import LegalPage from "@/components/LegalPage";
import { CONTACT_EMAIL, OPERATOR, SITE } from "@/lib/legal";

export const metadata: Metadata = {
  title: "Privacy Policy — smistudy",
  description: "How smistudy collects, uses, stores and protects your personal information.",
};

const mail = <a href={`mailto:${CONTACT_EMAIL}`}>{CONTACT_EMAIL}</a>;

export default function PrivacyPage() {
  return (
    <LegalPage
      title="Privacy Policy"
      summary={
        <p>
          <strong>The short version:</strong> we collect only what&apos;s needed to run your study tracker: your email, username,
          the study time you log and anything you write in notes. We don&apos;t sell your data, show ads, or use analytics or
          tracking cookies. Your profile is private unless you make it public. You can download or delete everything at any
          time from Settings.
        </p>
      }
    >
      <p>
        This Privacy Policy explains how smistudy (&quot;smistudy&quot;, &quot;we&quot;, &quot;us&quot;), available at {SITE},
        collects, uses, discloses and protects your personal information. It is written to meet Canada&apos;s{" "}
        <em>Personal Information Protection and Electronic Documents Act</em> (PIPEDA) and, for users in Quebec, the{" "}
        <em>Act respecting the protection of personal information in the private sector</em> as amended by Law 25.
      </p>

      <h2>1. Who is responsible for your information</h2>
      <p>
        smistudy is operated by {OPERATOR}, an individual based in Ontario, Canada. {OPERATOR} is accountable for the personal
        information we handle and is the person in charge of the protection of personal information for the purposes of
        Quebec law. You can reach them about anything in this policy at {mail}.
      </p>

      <h2>2. What we collect</h2>
      <h3>Information you give us</h3>
      <ul>
        <li><strong>Account details:</strong> your email address, username and display name.</li>
        <li>
          <strong>Password:</strong> if you sign up with email, we store only a one-way cryptographic hash of your password
          (Argon2id). We never store or see your actual password.
        </li>
        <li>
          <strong>Study data:</strong> the study sessions you log: the date, the length in minutes, whether it came from the
          Pomodoro timer or was logged manually, any note you add, and when it was recorded.
        </li>
        <li>
          <strong>To-do list:</strong> the tasks you add, whether each is done, and when it was created and completed. Your
          to-do list is private and never appears on your public profile or stats card.
        </li>
        <li>
          <strong>Reward goals:</strong> the rewards you set for yourself, their hour targets and start dates, and when you
          claimed them. These are private and never appear on your public profile or stats card.
        </li>
        <li><strong>Messages:</strong> anything you send us by email, such as a privacy request.</li>
      </ul>
      <h3>Information from Google, if you use &quot;Continue with Google&quot;</h3>
      <p>
        Google shares your Google account ID, email address (and whether Google has verified it) and your name. We use these
        only to create and sign in to your account. We don&apos;t receive your Google password, contacts or any other Google
        data.
      </p>
      <h3>Information collected automatically</h3>
      <ul>
        <li>
          <strong>Sign-in session:</strong> a random session token in a cookie (see our <Link href="/cookies">Cookies &amp;
          Storage notice</Link>). We store only a hash of it, along with your browser&apos;s user-agent string (e.g. &quot;Safari
          on macOS&quot;) and when the session started and expires.
        </li>
        <li>
          <strong>Security information:</strong> the number of recent failed log-in attempts on your account, and, briefly
          and in memory only, your IP address, used to limit how fast requests can be made so the site can&apos;t be abused.
          We don&apos;t store IP addresses in our database or tie them to your study data.
        </li>
        <li>
          <strong>Preferences on your device:</strong> your light/dark theme and Pomodoro timer lengths are saved in your
          browser&apos;s local storage. They never leave your device.
        </li>
      </ul>
      <p>
        We do <strong>not</strong> use analytics, advertising, social-media pixels, fingerprinting or any other tracking
        technology, and we don&apos;t build profiles about you or make automated decisions about you.
      </p>

      <h2>3. Why we use it</h2>
      <ul>
        <li>To create your account, sign you in and keep your account secure.</li>
        <li>To store your study sessions, to-do list and reward goals, and show you your graph, streaks, statistics and goal progress.</li>
        <li>To show your public profile, only if you choose to turn it on.</li>
        <li>
          To send emails about your account: confirming your email address, resetting your password, or telling you someone
          tried to sign up with your email. We don&apos;t send marketing or promotional emails.
        </li>
        <li>To prevent abuse, investigate security problems, and keep the service working.</li>
        <li>To respond to your questions and requests, and to meet legal obligations.</li>
      </ul>
      <p>We only use your information for these purposes. If we ever want to use it for something new, we&apos;ll ask you first.</p>

      <h2>4. Your consent</h2>
      <p>
        By creating an account, you consent to us collecting and using your information as described here. Making your profile
        public is a separate, opt-in choice (off by default) that you can turn off at any time in Settings. You can withdraw
        your consent at any time by deleting your account. Since we can&apos;t provide the service without the basic account
        and study information, deleting your account is how you withdraw consent to it.
      </p>

      <h2>5. Public profiles</h2>
      <p>
        Your profile is <strong>private by default</strong>. If you turn on &quot;Public profile&quot; in Settings, anyone with the
        link (<code>{SITE}/u/your-username</code>) can see your display name, username, the month you joined, your total and
        weekly study time, your streaks, and your study graph (minutes studied per day). Your email address, your notes and
        the details of individual sessions are <strong>never</strong> public. Turning the setting off hides your profile
        immediately.
      </p>

      <h2>6. Who we share it with</h2>
      <p>
        We don&apos;t sell, rent or trade your personal information. We share it only with these service providers, and only
        what each needs to do its job for us:
      </p>
      <table>
        <thead>
          <tr><th>Provider</th><th>What they do</th><th>What they receive</th><th>Location</th></tr>
        </thead>
        <tbody>
          <tr><td>OVHcloud</td><td>Hosts our server and database</td><td>All data stored by smistudy</td><td>France</td></tr>
          <tr>
            <td>Cloudflare</td>
            <td>Secures and delivers the website (protection against attacks)</td>
            <td>Your IP address and the web traffic between you and smistudy</td>
            <td>Global network, including the United States</td>
          </tr>
          <tr><td>Resend</td><td>Sends our account emails</td><td>Your email address and the email&apos;s contents</td><td>United States</td></tr>
          <tr>
            <td>Google</td>
            <td>Sign-in, only if you choose &quot;Continue with Google&quot;</td>
            <td>Google already has your account; it learns you signed in to smistudy</td>
            <td>United States</td>
          </tr>
        </tbody>
      </table>
      <p>
        We may also disclose information if required by law (for example, a valid court order), to protect the rights, safety
        or security of our users or the public, or as part of a transfer of the service to a new operator, who would be bound
        by this policy.
      </p>

      <h2>7. Where your information is stored</h2>
      <p>
        Your information is stored on a server in <strong>France</strong>, and some of our service providers process it in the
        <strong> United States</strong> and other countries. While it&apos;s outside Canada, it is protected by the laws of those
        countries (in France, the EU&apos;s GDPR) and may be accessible to their courts and authorities. Before using these
        providers we assessed their privacy and security practices, and we use them only on terms that require them to protect
        your information.
      </p>

      <h2>8. How long we keep it</h2>
      <ul>
        <li><strong>Account and study data:</strong> as long as your account exists.</li>
        <li>
          <strong>When you delete your account:</strong> your account, sessions, study data, to-do list and reward goals are removed from our live
          database immediately. Copies in our daily backups are deleted automatically within 14 days.
        </li>
        <li><strong>Sign-in sessions:</strong> expire 30 days after you were last active (sooner if you log out).</li>
        <li><strong>Email links:</strong> confirmation links expire after 24 hours; password-reset links after 1 hour.</li>
        <li><strong>IP addresses for rate limiting:</strong> held in memory for minutes, never written to our database.</li>
        <li>Emails sent through Resend are kept in their logs according to their own retention settings.</li>
      </ul>

      <h2>9. How we protect it</h2>
      <p>
        We use safeguards appropriate to the sensitivity of the information. These include encrypted connections (HTTPS)
        everywhere; strongly hashed passwords and session tokens; secure, HttpOnly cookies; protection against
        cross-site request forgery; rate limiting and account lockout against password guessing; locked-down, sandboxed server
        processes; a firewall; and restricted access to the server. No system is perfectly secure, but we work to keep your
        information safe and fix problems quickly.
      </p>

      <h2>10. Your rights</h2>
      <p>You have the right to:</p>
      <ul>
        <li>
          <strong>Access</strong> your information and receive a copy in a common, structured format: use{" "}
          <strong>Settings → Download my data</strong>, or email us.
        </li>
        <li><strong>Correct</strong> it: edit your username and display name in Settings, or email us about anything else.</li>
        <li><strong>Delete</strong> it: use <strong>Settings → Delete account</strong>.</li>
        <li><strong>Withdraw consent</strong> or turn your public profile off at any time.</li>
        <li>Ask how your information has been used and to whom it has been disclosed.</li>
      </ul>
      <p>
        Email {mail} for any request. We&apos;ll respond within 30 days and may need to confirm your identity first. Requests are
        free. If you&apos;re not satisfied with our response, you can complain to the{" "}
        <a href="https://www.priv.gc.ca" target="_blank" rel="noreferrer">Office of the Privacy Commissioner of Canada</a> or,
        if you live in Quebec, the{" "}
        <a href="https://www.cai.gouv.qc.ca" target="_blank" rel="noreferrer">Commission d&apos;accès à l&apos;information du Québec</a>.
      </p>

      <h2>11. If something goes wrong</h2>
      <p>
        If a breach of security involving your personal information creates a real risk of significant harm to you, we will
        notify you as soon as feasible, report it to the Office of the Privacy Commissioner of Canada (and, for Quebec
        residents, the Commission d&apos;accès à l&apos;information), and keep a record of it, as the law requires.
      </p>

      <h2>12. Children</h2>
      <p>
        smistudy is for people aged <strong>13 and older</strong>. If you&apos;re under the age of majority where you live, you
        need a parent or guardian&apos;s permission to use smistudy. In Quebec, users under 14 need a parent&apos;s or
        guardian&apos;s consent to the collection of their personal information. If we learn that a child under 13 has created
        an account, we&apos;ll delete it. Parents or guardians can contact us at {mail}.
      </p>

      <h2>13. Changes to this policy</h2>
      <p>
        If we change this policy, we&apos;ll update the date at the top. If a change materially affects how we use your
        information, we&apos;ll tell you by email or on the site before it takes effect, and ask for your consent where the law
        requires it.
      </p>

      <h2>14. Contact</h2>
      <p>
        Questions, requests or complaints: {mail}. Attention: {OPERATOR}, person in charge of the protection of personal
        information, smistudy.
      </p>
    </LegalPage>
  );
}

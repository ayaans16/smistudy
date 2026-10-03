import type { Metadata } from "next";
import Link from "next/link";
import LegalPage from "@/components/LegalPage";
import { CONTACT_EMAIL, OPERATOR, SITE } from "@/lib/legal";

export const metadata: Metadata = {
  title: "Terms of Service — smistudy",
  description: "The terms for using smistudy.",
};

const mail = <a href={`mailto:${CONTACT_EMAIL}`}>{CONTACT_EMAIL}</a>;

export default function TermsPage() {
  return (
    <LegalPage
      title="Terms of Service"
      summary={
        <p>
          <strong>The short version:</strong> smistudy is a free study tracker run by one person. Be kind, don&apos;t
          abuse the site, and keep your account secure. Your study data is yours. We provide the service as-is and may change
          it, but we&apos;ll never take away rights that consumer law gives you.
        </p>
      }
    >
      <p>
        These Terms of Service (&quot;Terms&quot;) are an agreement between you and {OPERATOR} (&quot;we&quot;, &quot;us&quot;),
        who operates smistudy at {SITE} (the &quot;Service&quot;). By creating an account or using the Service, you agree to
        these Terms and to our <Link href="/privacy">Privacy Policy</Link>. If you don&apos;t agree, please don&apos;t use the
        Service.
      </p>

      <h2>1. Who can use smistudy</h2>
      <p>
        You must be at least <strong>13 years old</strong>. If you&apos;re under the age of majority where you live, you need
        a parent or guardian&apos;s permission, and they agree to these Terms on your behalf. In Quebec, users under 14 need a
        parent&apos;s or guardian&apos;s consent to the collection of their personal information.
      </p>

      <h2>2. Your account</h2>
      <ul>
        <li>Give accurate information when you sign up, and keep your email address up to date.</li>
        <li>
          Keep your password secret and don&apos;t share your account. You&apos;re responsible for activity on your account
          unless it results from our failure to keep the Service secure.
        </li>
        <li>Tell us right away at {mail} if you think someone else has accessed your account.</li>
        <li>One person per account. Don&apos;t create accounts in bulk or by automated means.</li>
      </ul>

      <h2>3. The Service</h2>
      <p>
        smistudy is currently free. It provides a Pomodoro timer, a study-time tracker and optional public profiles. We may add,
        change or remove features, and may suspend or discontinue the Service. If we plan to shut it down, we&apos;ll give you
        reasonable notice so you can download your data.
      </p>

      <h2>4. Your content</h2>
      <p>
        &quot;Your content&quot; means the study sessions, notes, username and display name you add. You own your content. You
        give us a limited, non-exclusive, royalty-free licence to store, process and display it <em>only</em> as needed to
        run the Service for you, including showing your public profile if you turn it on. This licence ends when you delete
        the content or your account, except for copies in backups that are deleted on schedule (see the Privacy Policy).
      </p>
      <p>
        You&apos;re responsible for your content. Your username and display name may be public, so they must not be offensive,
        hateful or misleading, and must not impersonate anyone or infringe anyone&apos;s rights.
      </p>

      <h2>5. Acceptable use</h2>
      <p>You agree not to:</p>
      <ul>
        <li>break any law, or use the Service to harass, threaten, impersonate or harm anyone;</li>
        <li>access or try to access other people&apos;s accounts or data;</li>
        <li>
          attack, probe, overload or disrupt the Service. This includes getting around rate limits, security measures or
          access controls, and introducing malware;
        </li>
        <li>scrape, crawl or collect data from the Service by automated means, except public profiles at a reasonable rate;</li>
        <li>use the Service to send spam or to build a competing product from our code or design.</li>
      </ul>
      <p>
        <strong>Found a security issue?</strong> Please report it to {mail} and give us a reasonable chance to fix it before
        disclosing it. We appreciate good-faith research that doesn&apos;t access other people&apos;s data or disrupt the
        Service.
      </p>

      <h2>6. Fan project and trademarks</h2>
      <p>
        smistudy is an independent, fan-made project. It is <strong>not affiliated with, sponsored by or endorsed by Dreams
        Inc.</strong>, and &quot;Smiski&quot; and &quot;Sonny Angel&quot; are trademarks of their respective owners, used here
        only to describe the inspiration for the site&apos;s style. The mascot illustrations on smistudy are original drawings,
        not official artwork.
      </p>

      <h2>7. Open-source code</h2>
      <p>
        smistudy&apos;s source code, including its original illustrations, is open source under the{" "}
        <a href="https://www.gnu.org/licenses/agpl-3.0.html" target="_blank" rel="noreferrer">GNU Affero General Public License
        v3.0</a> and is available at{" "}
        <a href="https://github.com/ayaans16/smistudy" target="_blank" rel="noreferrer">github.com/ayaans16/smistudy</a>. You may
        use, change and share it under that licence. These Terms don&apos;t restrict any rights the licence gives you, but if
        you run your own copy, it isn&apos;t smistudy: please use a different name, and don&apos;t suggest that we run or
        endorse it.
      </p>

      <h2>8. Suspension and termination</h2>
      <p>
        You can stop using smistudy and delete your account at any time in Settings. We may suspend or close an account that
        seriously or repeatedly breaks these Terms, or where needed to protect the Service or other users. Where reasonable,
        we&apos;ll tell you why first and give you a chance to download your data. Sections 4 (for content already removed),
        6, 7, 9, 10 and 12 continue to apply after your account ends.
      </p>

      <h2>9. Disclaimer</h2>
      <p>
        We work hard to keep smistudy reliable, but it&apos;s provided <strong>&quot;as is&quot; and &quot;as available&quot;</strong>.
        To the extent permitted by law, we make no promises that it will be uninterrupted, error-free or that data will never be
        lost. We back up data daily, but keeping your own records is wise. smistudy is a productivity tool and does not provide
        academic, medical or professional advice.
      </p>

      <h2>10. Limitation of liability</h2>
      <p>
        To the extent permitted by law, we are not liable for indirect, incidental, special or consequential damages, or for
        lost data, profits or opportunities, arising from your use of the Service. Our total liability for any claim related
        to the Service is limited to CAD $50. Nothing in these Terms excludes or limits liability that can&apos;t be excluded
        under applicable law, including liability for gross negligence or intentional misconduct.
      </p>

      <h2>11. Your consumer rights</h2>
      <p>
        Nothing in these Terms takes away rights you have under consumer protection laws that can&apos;t be waived by contract,
        including Ontario&apos;s <em>Consumer Protection Act, 2002</em> and Quebec&apos;s <em>Consumer Protection Act</em>. If you
        are a consumer in Quebec, any clause in these Terms that is prohibited or unenforceable under Quebec law (for example,
        limitations of liability or warranty exclusions) does not apply to you.
      </p>

      <h2>12. Governing law and disputes</h2>
      <p>
        These Terms are governed by the laws of the Province of Ontario and the federal laws of Canada that apply there. Disputes
        will be handled by the courts of Ontario, except that if you are a consumer, you may also bring proceedings in the courts
        of the province or territory where you live, and you keep the protection of its mandatory laws. Before going to court,
        please contact us at {mail}. Most problems can be sorted out quickly.
      </p>

      <h2>13. Changes to these Terms</h2>
      <p>
        We may update these Terms. For material changes, we&apos;ll notify you by email or on the site at least 30 days before
        they take effect, explaining what&apos;s changing. If you don&apos;t agree, you can delete your account before then.
        Continuing to use the Service after the changes take effect means you accept them.
      </p>

      <h2>14. General</h2>
      <p>
        If any part of these Terms is found unenforceable, the rest stays in effect. Our not enforcing a provision isn&apos;t a
        waiver of it. These Terms and the Privacy Policy are the entire agreement between you and us about the Service. You
        may not transfer your rights under these Terms. We may transfer them to someone who takes over the Service, who will be
        bound by them.
      </p>

      <h2>15. Contact</h2>
      <p>Questions about these Terms: {mail}.</p>
    </LegalPage>
  );
}

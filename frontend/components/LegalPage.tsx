import Link from "next/link";
import { LAST_UPDATED } from "@/lib/legal";
import SiteHeader from "./SiteHeader";

const otherPages = [
  { href: "/terms", label: "Terms of Service" },
  { href: "/privacy", label: "Privacy Policy" },
  { href: "/cookies", label: "Cookies & Storage" },
];

/** Readable long-form layout shared by the Terms, Privacy and Cookies pages. */
export default function LegalPage({ title, summary, children }: { title: string; summary: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="mx-auto w-full max-w-6xl px-4 pb-16 sm:px-6">
      <SiteHeader />
      <main className="mx-auto max-w-3xl">
        <nav className="mb-6 flex flex-wrap gap-2 text-sm" aria-label="Legal pages">
          {otherPages.map((p) => (
            <Link key={p.href} href={p.href} className="rounded-full border border-line px-3 py-1 font-semibold text-muted transition hover:border-accent hover:text-fg">
              {p.label}
            </Link>
          ))}
        </nav>
        <article className="rounded-3xl border border-line bg-card p-6 shadow-sm sm:p-10 [&_a]:font-semibold [&_a]:text-accent-strong [&_a:hover]:underline [&_h2]:mt-10 [&_h2]:text-xl [&_h2]:font-extrabold [&_h3]:mt-6 [&_h3]:font-bold [&_li]:mt-1.5 [&_p]:mt-3 [&_p]:leading-relaxed [&_table]:mt-4 [&_table]:w-full [&_table]:text-left [&_table]:text-sm [&_td]:border-t [&_td]:border-line [&_td]:py-2 [&_td]:pr-3 [&_td]:align-top [&_td]:[overflow-wrap:anywhere] [&_th]:pb-2 [&_th]:pr-3 [&_th]:text-xs [&_th]:uppercase [&_th]:tracking-wide [&_th]:text-muted [&_ul]:mt-3 [&_ul]:list-disc [&_ul]:space-y-1 [&_ul]:pl-6">
          <h1 className="text-3xl font-extrabold">{title}</h1>
          <p className="!mt-1 text-sm text-muted">Last updated: {LAST_UPDATED}</p>
          <div className="mt-6 rounded-2xl bg-accent-soft p-4 text-sm [&_p]:!mt-0">{summary}</div>
          {children}
        </article>
      </main>
    </div>
  );
}

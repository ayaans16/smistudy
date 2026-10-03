import Link from "next/link";

export default function SiteFooter() {
  return (
    <footer className="mx-auto mt-auto w-full max-w-6xl px-4 py-8 text-center text-xs text-muted sm:px-6">
      <nav className="flex flex-wrap justify-center gap-x-4 gap-y-1" aria-label="Legal">
        <Link href="/terms" className="hover:text-fg">Terms</Link>
        <Link href="/privacy" className="hover:text-fg">Privacy</Link>
        <Link href="/cookies" className="hover:text-fg">Cookies</Link>
      </nav>
      <p className="mt-2">
        © {new Date().getFullYear()} smistudy · A fan-made project, not affiliated with or endorsed by Dreams Inc., the makers of
        Smiski and Sonny Angel. Mascots are original drawings.
      </p>
    </footer>
  );
}

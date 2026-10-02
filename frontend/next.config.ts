import type { NextConfig } from "next";

const backend = process.env.SMISTUDY_API_URL ?? "http://localhost:8080";

// Next and the theme script use inline <script>s, hence 'unsafe-inline'; everything
// else is locked to this origin. Dev mode needs eval, so headers are production-only.
const contentSecurityPolicy = [
  "default-src 'self'",
  "script-src 'self' 'unsafe-inline'",
  "style-src 'self' 'unsafe-inline'",
  "img-src 'self' data:",
  "font-src 'self'",
  "connect-src 'self'",
  "object-src 'none'",
  "base-uri 'self'",
  "form-action 'self'",
  "frame-ancestors 'none'",
].join("; ");

const securityHeaders = [
  { key: "Content-Security-Policy", value: contentSecurityPolicy },
  { key: "X-Frame-Options", value: "DENY" },
  { key: "X-Content-Type-Options", value: "nosniff" },
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=(), payment=()" },
];

const nextConfig: NextConfig = {
  // Self-contained server build for the Docker image; plain `next start` everywhere else.
  output: process.env.NEXT_STANDALONE ? "standalone" : undefined,
  poweredByHeader: false,
  // Proxy API calls to the Go backend so the browser only talks to one origin.
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${backend}/api/:path*` }];
  },
  async headers() {
    return process.env.NODE_ENV === "production" ? [{ source: "/:path*", headers: securityHeaders }] : [];
  },
};

export default nextConfig;

import type { NextConfig } from "next";

const backend = process.env.SMISTUDY_API_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  // Self-contained server build for the Docker image.
  output: "standalone",
  // Proxy API calls to the Go backend so the browser only talks to one origin.
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${backend}/api/:path*` }];
  },
};

export default nextConfig;

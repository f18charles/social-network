"use strict";

import { resolveApiUrl } from "./api";

const PING_INTERVAL_MS = 10 * 60 * 1000; // 10 minutes

/**
 * Pings the backend health endpoint once.
 *
 * Uses a plain fetch (no credentials required) so the request works even
 * before the user has authenticated and avoids triggering any CORS
 * preflight for a cookie-bearing request.
 *
 * @returns {Promise<void>}
 */
const ping = async () => {
  const url = resolveApiUrl("/health");
  try {
    await fetch(url, { method: "GET", credentials: "omit" });
  } catch {
    // Network errors are silently swallowed — the backend is just
    // warming up or temporarily unavailable; we'll retry next interval.
  }
};

/**
 * Starts a recurring ping against the backend /health endpoint so that
 * Render's free-tier service does not spin down between user sessions.
 *
 * @returns {() => void} A cleanup function that stops the pinger when called.
 */
export const startPingService = () => {
  // Fire once immediately so the backend is warm on first load.
  ping();

  const intervalId = setInterval(ping, PING_INTERVAL_MS);

  return () => clearInterval(intervalId);
};

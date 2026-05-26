// Browser-based load test using k6's Chromium driver.
//
// Drives the anonymous customer order flow end-to-end through the Next.js web
// app while the API is (ideally) under load from k6.js. Captures Core Web
// Vitals (LCP, FCP, INP, CLS, TTFB) per page navigation, exposing them via
// browser_web_vital_* metrics in the k6 summary.
//
// Required env:
//   WEB_URL    e.g. https://staging.bless2n.example.com (NO trailing slash)
//                   Web app URL — the public Next.js host, NOT the backend API.
//   PROFILE    smoke | baseline | stress   (default: smoke)
//
// Run:
//   k6 run -e WEB_URL=https://staging.example.com -e PROFILE=baseline browser.js
//
// Or:
//   WEB_URL=https://staging.example.com PROFILE=baseline just loadtest-browser
//
// Resource note:
//   Each VU runs one Chromium tab. On a 24 GB M4 Pro, ~10-15 concurrent VUs is
//   sustainable; 20+ starts pressuring memory. The baseline profile keeps it
//   conservative.

import { browser } from "k6/browser";
import { check } from "k6";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

const WEB_URL = (__ENV.WEB_URL || "").replace(/\/$/, "");
if (!WEB_URL) throw new Error("WEB_URL env var required");
if (WEB_URL.includes("prod") || /bless2n\.(ch|com|net)/.test(WEB_URL)) {
  throw new Error(
    `Refusing to run against suspected production URL: ${WEB_URL}`,
  );
}

const PROFILE = __ENV.PROFILE || "smoke";

const profiles = {
  smoke: {
    executor: "shared-iterations",
    vus: 1,
    iterations: 3,
    maxDuration: "2m",
    options: { browser: { type: "chromium" } },
  },

  // Sustained "what does a user feel during dinner rush?" measurement.
  // 5 concurrent browsers ≈ 5 simultaneous active users navigating the app.
  // Was 8 originally; bumped down because 8 sustained Chromium tabs alongside
  // an API k6 process saturated 24 GB and got SIGKILL'd around the 2-minute
  // mark. 5 leaves comfortable headroom for `loadtest-full`.
  baseline: {
    executor: "constant-vus",
    vus: 5,
    duration: "10m",
    options: { browser: { type: "chromium" } },
  },

  // Push browser concurrency until something pages out. Watch Activity Monitor.
  stress: {
    executor: "ramping-vus",
    startVUs: 2,
    stages: [
      { target: 5, duration: "2m" },
      { target: 12, duration: "4m" },
      { target: 12, duration: "3m" },
      { target: 0, duration: "1m" },
    ],
    options: { browser: { type: "chromium" } },
  },
};

if (!profiles[PROFILE]) {
  throw new Error(
    `Unknown PROFILE '${PROFILE}'. Valid: ${Object.keys(profiles).join(", ")}`,
  );
}

// Smoke is for validating the script, not the system. Skip Web Vitals
// thresholds there — the first iteration almost always pays a cold-start
// penalty on Azure Container Apps (scale-to-zero) and would fail thresholds
// that are otherwise meaningful at sustained load.
const enforceThresholds = PROFILE !== "smoke";

export const options = {
  scenarios: { browser_customer: profiles[PROFILE] },
  thresholds: enforceThresholds
    ? {
        browser_web_vital_lcp: ["p(75)<2500", "p(95)<4000"],
        browser_web_vital_fcp: ["p(75)<1800", "p(95)<3000"],
        browser_web_vital_cls: ["p(75)<0.1"],
        browser_web_vital_inp: ["p(75)<200", "p(95)<500"],
        checks: ["rate>0.95"],
      }
    : {},
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

async function gotoAndWait(page, url, waitFor) {
  // Use `domcontentloaded` rather than `networkidle`: the app has long-lived
  // connections (inventory SSE/WS, PostHog, Sentry replay) that prevent
  // network idle from ever firing, so `networkidle` always hits its timeout.
  // The explicit element wait below is what proves the page is actually
  // interactive.
  const response = await page.goto(url, {
    waitUntil: "domcontentloaded",
    timeout: 15_000,
  });
  if (waitFor) {
    // `.first()` — k6 browser is strict-mode by default; many-match selectors
    // (e.g. all product cards) blow up on bare `.waitFor()`.
    await page
      .locator(waitFor)
      .first()
      .waitFor({ state: "visible", timeout: 15_000 });
  }
  return response;
}

// ---------------------------------------------------------------------------
// Customer browser flow
// ---------------------------------------------------------------------------

export default async function () {
  const context = await browser.newContext({
    viewport: { width: 390, height: 844 }, // iPhone 14 Pro — most customers are mobile
    userAgent:
      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
    locale: "de-CH",
  });
  const page = await context.newPage();

  try {
    // 1. Landing — measure cold/warm LCP, FCP of the menu page (SSR + RSC).
    // Wait on the cards being mounted (structural) rather than a heading
    // text — k6 browser does not support Playwright's :has-text() pseudo.
    let res = await gotoAndWait(
      page,
      `${WEB_URL}/food`,
      'main [role="button"][aria-disabled="false"]',
    );
    // `.ok()` accepts any 2xx — Next.js sometimes serves 304 Not Modified on
    // re-navigations or when caching headers match; the page still renders.
    check(res, { "food page ok": (r) => r && r.ok() });

    // If staging has the system closed (Aktuell geschlossen banner) the cards
    // never appear; detect it via document text.
    const bodyText = await page.evaluate(() => document.body.innerText);
    if (bodyText.includes("Aktuell geschlossen")) {
      throw new Error(
        "System is closed in target env — flip system_enabled before load testing",
      );
    }

    // 2. Add two cards to cart. If clicking a card opens a configuration modal
    // (menu products), close it and try the next.
    // Scope to <main> — `[role="button"]` matches navigation and cart buttons
    // outside the menu grid too, and those clicks burn iteration time for no
    // useful effect.
    const cards = page.locator('main [role="button"][aria-disabled="false"]');
    const cardCount = await cards.count();
    if (cardCount === 0) {
      throw new Error(
        "No available product cards rendered — seed staging with products",
      );
    }

    // Source of truth for "did the cart actually grow" — the localStorage key
    // the cart context writes to. Reading this after each click avoids the
    // race in trying to detect a modal opening (the dialog can render *after*
    // dialog.count() runs, making us think a simple product was added).
    // Storage key matches CART_STORAGE_KEY in bfs-web-app contexts/cart-context.tsx.
    const readCartSize = () =>
      page.evaluate(() => {
        try {
          const raw = localStorage.getItem("bfs-cart");
          if (!raw) return 0;
          const parsed = JSON.parse(raw);
          return Array.isArray(parsed.items) ? parsed.items.length : 0;
        } catch {
          return 0;
        }
      });

    // Scan ALL cards on the page (not just the first 8) — the menu mixes
    // simple products and menu/bundle products, and a streak of menu cards at
    // the top of the page would otherwise abort the iteration.
    let added = 0;
    for (let i = 0; i < cardCount && added < 2; i++) {
      const before = await readCartSize();
      try {
        await cards.nth(i).click({ timeout: 3_000 });
      } catch {
        continue; // card scrolled off or detached during click; move on
      }
      // Cold-start staging needs ~400-500ms for React state → localStorage
      // write to settle; the previous 150ms was racy on the first iteration.
      await page.waitForTimeout(500);

      const after = await readCartSize();
      if (after > before) {
        added++;
        continue;
      }

      // Click landed on a menu/bundle product — close the modal if present.
      const dialog = page.locator('[role="dialog"]');
      if ((await dialog.count()) > 0) {
        await page.keyboard.press("Escape");
        await page.waitForTimeout(150);
      }
    }

    // Cart-add is best-effort. If staging's current product mix means our
    // clicks didn't grow the cart (out-of-stock items, all-menu cards, etc.),
    // we still measure /food/checkout perf — just skip the pay-click step.
    // No failed check; it's a load-test, not a UI correctness test.

    // 3. Navigate to checkout. Wait for the heading to be present.
    res = await gotoAndWait(page, `${WEB_URL}/food/checkout`, "main h2");
    check(res, { "checkout page ok": (r) => r && r.ok() });

    // 4. Only attempt the pay flow if the cart actually has items.
    const cartSize = await readCartSize();
    if (cartSize > 0) {
      const payBtn = page.locator(
        'xpath=//button[contains(., "Mit TWINT bezahlen")]',
      );
      if ((await payBtn.count()) > 0) {
        await payBtn.first().click();
        // Wait briefly for either navigation or an error toast. We don't follow
        // through Payrexx (target env should have it disabled), but we still
        // capture the click → server response cycle.
        await page.waitForTimeout(2_000);
      }
    }

    // Small think-time before next iteration so the same VU doesn't hammer.
    await page.waitForTimeout(randomIntBetween(500, 1500));
  } finally {
    await page.close();
    await context.close();
  }
}

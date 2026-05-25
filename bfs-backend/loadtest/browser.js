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
  // 8 concurrent browsers ≈ 8 simultaneous active users navigating the app.
  baseline: {
    executor: "constant-vus",
    vus: 8,
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

export const options = {
  scenarios: { browser_customer: profiles[PROFILE] },
  thresholds: {
    // Core Web Vitals targets — tighten as the system improves.
    browser_web_vital_lcp: ["p(75)<2500", "p(95)<4000"],
    browser_web_vital_fcp: ["p(75)<1800", "p(95)<3000"],
    browser_web_vital_cls: ["p(75)<0.1"],
    browser_web_vital_inp: ["p(75)<200", "p(95)<500"],
    checks: ["rate>0.95"],
  },
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

async function gotoAndWait(page, url, waitFor) {
  const response = await page.goto(url, {
    waitUntil: "networkidle",
    timeout: 30_000,
  });
  if (waitFor) {
    await page.locator(waitFor).waitFor({ state: "visible", timeout: 15_000 });
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
    let res = await gotoAndWait(
      page,
      `${WEB_URL}/food`,
      'h2:has-text("Alle Produkte")',
    );
    check(res, { "food page 200": (r) => r && r.status() === 200 });

    // If staging has the system closed (Aktuell geschlossen banner), abort.
    const closedBanner = page.locator('h2:has-text("Aktuell geschlossen")');
    if ((await closedBanner.count()) > 0) {
      throw new Error(
        "System is closed in target env — flip system_enabled before load testing",
      );
    }

    // 2. Wait for product cards to hydrate, then add two to cart.
    const cards = page.locator('[role="button"][aria-disabled="false"]');
    const cardCount = await cards.count();
    if (cardCount === 0) {
      throw new Error(
        "No available product cards rendered — seed staging with products",
      );
    }

    // Click two distinct cards. If clicking a card opens a configuration
    // modal (menu products), close it and try the next.
    let added = 0;
    for (let i = 0; i < Math.min(cardCount, 6) && added < 2; i++) {
      const card = cards.nth(i);
      await card.click({ timeout: 5_000 });

      // If a modal opened (menu/bundle product), bail and try the next card.
      const dialog = page.locator('[role="dialog"]');
      if ((await dialog.count()) > 0) {
        await page.keyboard.press("Escape");
        await page.waitForTimeout(200);
        continue;
      }
      added++;
    }

    if (added === 0) {
      throw new Error(
        "Failed to add any simple product to cart (all visible cards open modals)",
      );
    }

    // 3. Navigate to checkout. Captures the LCP / FCP / INP of /food/checkout.
    res = await gotoAndWait(
      page,
      `${WEB_URL}/food/checkout`,
      'h2:has-text("Warenkorb")',
    );
    check(res, { "checkout page 200": (r) => r && r.status() === 200 });

    // 4. Verify cart isn't empty (i.e. localStorage persisted across goto).
    const emptyCart = page.locator('p:has-text("Warenkorb ist leer")');
    check(null, {
      "cart populated": () => emptyCart.count().then((c) => c === 0),
    });

    // 5. Click "Mit TWINT bezahlen". This triggers POST /v1/orders + payment
    // flow on the backend — the most expensive synchronous path in the system.
    const payBtn = page.locator('button:has-text("Mit TWINT bezahlen")');
    if ((await payBtn.count()) > 0) {
      await payBtn.click();
      // Wait briefly for either navigation or an error toast. We don't follow
      // through Payrexx (target env should have it disabled), but we still
      // capture the click → server response cycle.
      await page.waitForTimeout(2_000);
    }

    // Small think-time before next iteration so the same VU doesn't hammer.
    await page.waitForTimeout(randomIntBetween(500, 1500));
  } finally {
    await page.close();
    await context.close();
  }
}

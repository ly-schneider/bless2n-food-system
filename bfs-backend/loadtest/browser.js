// Browser-based load test driving the anonymous customer order flow through
// the Next.js web app, capturing Core Web Vitals per page navigation.
//
// Env: WEB_URL (public web host, not API), PROFILE = smoke | baseline | stress
//
// Each VU runs one Chromium tab (~200 MB). The baseline profile keeps
// concurrency conservative for 24 GB machines.

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
  baseline: {
    executor: "constant-vus",
    vus: 5,
    duration: "10m",
    options: { browser: { type: "chromium" } },
  },
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

// Smoke validates the script; cold-start would skew Web Vitals thresholds.
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

async function gotoAndWait(page, url, waitFor) {
  // `networkidle` never fires — inventory SSE + PostHog + Sentry keep
  // connections open indefinitely.
  const response = await page.goto(url, {
    waitUntil: "domcontentloaded",
    timeout: 15_000,
  });
  if (waitFor) {
    // k6 browser is strict-mode by default; many-match selectors need .first().
    await page
      .locator(waitFor)
      .first()
      .waitFor({ state: "visible", timeout: 15_000 });
  }
  return response;
}

export default async function () {
  const context = await browser.newContext({
    viewport: { width: 390, height: 844 },
    userAgent:
      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
    locale: "de-CH",
  });
  const page = await context.newPage();

  try {
    let res = await gotoAndWait(
      page,
      `${WEB_URL}/food`,
      'main [role="button"][aria-disabled="false"]',
    );
    check(res, { "food page ok": (r) => r && r.ok() });

    const bodyText = await page.evaluate(() => document.body.innerText);
    if (bodyText.includes("Aktuell geschlossen")) {
      throw new Error(
        "System is closed in target env — flip system_enabled before load testing",
      );
    }

    // Scope to <main> — role="button" matches nav and cart buttons too.
    const cards = page.locator('main [role="button"][aria-disabled="false"]');
    const cardCount = await cards.count();
    if (cardCount === 0) {
      throw new Error(
        "No available product cards rendered — seed staging with products",
      );
    }

    // Storage key matches CART_STORAGE_KEY in bfs-web-app/contexts/cart-context.tsx.
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

    let added = 0;
    for (let i = 0; i < cardCount && added < 2; i++) {
      const before = await readCartSize();
      try {
        await cards.nth(i).click({ timeout: 3_000 });
      } catch {
        continue;
      }
      // 500ms covers React state → localStorage flush on a cold-started replica.
      await page.waitForTimeout(500);

      if ((await readCartSize()) > before) {
        added++;
        continue;
      }

      const dialog = page.locator('[role="dialog"]');
      if ((await dialog.count()) > 0) {
        await page.keyboard.press("Escape");
        await page.waitForTimeout(150);
      }
    }

    // Cart-add is best-effort: still measure /food/checkout perf even when
    // the click loop failed, just skip the pay step.
    res = await gotoAndWait(page, `${WEB_URL}/food/checkout`, "main h2");
    check(res, { "checkout page ok": (r) => r && r.ok() });

    if ((await readCartSize()) > 0) {
      // XPath because k6 browser doesn't support CSS :has-text().
      const payBtn = page.locator(
        'xpath=//button[contains(., "Mit TWINT bezahlen")]',
      );
      if ((await payBtn.count()) > 0) {
        await payBtn.first().click();
        await page.waitForTimeout(2_000);
      }
    }

    await page.waitForTimeout(randomIntBetween(500, 1500));
  } finally {
    await page.close();
    await context.close();
  }
}

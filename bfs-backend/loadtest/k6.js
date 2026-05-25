// k6 load test for the bless2n-food-system backend.
//
// Required env:
//   BASE_URL   e.g. https://api.staging.bless2n.example.com (NO trailing slash)
//   PROFILE    smoke | baseline | stress   (default: smoke)
//
// Run:
//   k6 run -e BASE_URL=https://api.staging.example.com -e PROFILE=baseline k6.js
//
// Or via just:
//   just loadtest BASE_URL=… PROFILE=baseline
//
// Prerequisites in the target environment:
//   1. Run seed.sql (see README) so the test tokens exist.
//   2. At least one approved STATION device with products assigned via device_product.
//   3. Payrexx NOT configured, so TWINT payments fall through to MarkOrderPaidDev.
//      Confirm by checking that h.payments.IsPayrexxEnabled() returns false in target env.

import http from "k6/http";
import { check, group, sleep } from "k6";
import { SharedArray } from "k6/data";
import {
  randomIntBetween,
  uuidv4,
} from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

const BASE_URL = (__ENV.BASE_URL || "").replace(/\/$/, "");
if (!BASE_URL) throw new Error("BASE_URL env var required");
if (BASE_URL.includes("prod") || BASE_URL.includes("api.bless2n")) {
  throw new Error(
    `Refusing to run against suspected production URL: ${BASE_URL}`,
  );
}

const PROFILE = __ENV.PROFILE || "smoke";

const tokens = JSON.parse(open("./tokens.json"));

// ---------------------------------------------------------------------------
// Scenario profiles
// ---------------------------------------------------------------------------
//
// `baseline` mirrors the dinner-rush shape seen in production around 18:30–22:30:
// fast ramp up, sustained peak, gradual cool down. Compressed to ~12 minutes
// for iteration speed; multiply durations to simulate a full evening.

const profiles = {
  smoke: {
    customer: {
      executor: "constant-vus",
      vus: 1,
      duration: "30s",
      tags: { scenario: "customer" },
    },
    admin: {
      executor: "constant-vus",
      vus: 1,
      duration: "30s",
      tags: { scenario: "admin" },
    },
    station: {
      executor: "constant-vus",
      vus: 1,
      duration: "30s",
      tags: { scenario: "station" },
    },
  },

  baseline: {
    customer: {
      executor: "ramping-arrival-rate",
      startRate: 1,
      timeUnit: "1s",
      preAllocatedVUs: 30,
      maxVUs: 80,
      stages: [
        { target: 2, duration: "1m" }, // pre-rush warmup (avoids cold-start measurement)
        { target: 7, duration: "2m" }, // ramp into rush
        { target: 7, duration: "5m" }, // sustained dinner peak
        { target: 3, duration: "2m" }, // post-rush trickle
        { target: 0, duration: "1m" }, // cool down
      ],
      tags: { scenario: "customer" },
    },
    admin: {
      executor: "constant-arrival-rate",
      rate: 1,
      timeUnit: "10s",
      preAllocatedVUs: 2,
      maxVUs: 5,
      duration: "11m",
      tags: { scenario: "admin" },
    },
    station: {
      executor: "ramping-arrival-rate",
      startRate: 1,
      timeUnit: "1s",
      preAllocatedVUs: 10,
      maxVUs: 30,
      stages: [
        { target: 1, duration: "1m" },
        { target: 3, duration: "2m" },
        { target: 3, duration: "5m" },
        { target: 1, duration: "2m" },
        { target: 0, duration: "1m" },
      ],
      tags: { scenario: "station" },
    },
  },

  stress: {
    customer: {
      executor: "ramping-arrival-rate",
      startRate: 1,
      timeUnit: "1s",
      preAllocatedVUs: 50,
      maxVUs: 200,
      stages: [
        { target: 5, duration: "1m" },
        { target: 20, duration: "3m" },
        { target: 50, duration: "5m" },
        { target: 0, duration: "1m" },
      ],
      tags: { scenario: "customer" },
    },
  },
};

if (!profiles[PROFILE]) {
  throw new Error(
    `Unknown PROFILE '${PROFILE}'. Valid: ${Object.keys(profiles).join(", ")}`,
  );
}

export const options = {
  scenarios: profiles[PROFILE],
  thresholds: {
    http_req_failed: ["rate<0.02"],
    "http_req_duration{scenario:customer}": ["p(95)<1500"],
    "http_req_duration{scenario:station}": ["p(95)<1000"],
    "http_req_duration{scenario:admin}": ["p(95)<2000"],
  },
  // Drop redirects for cleaner timings.
  maxRedirects: 0,
};

// ---------------------------------------------------------------------------
// Setup: prefetch the product catalog once. Shared by every VU via the return.
// ---------------------------------------------------------------------------

export function setup() {
  const res = http.get(`${BASE_URL}/v1/products`);
  if (res.status !== 200) {
    throw new Error(
      `setup: GET /v1/products returned ${res.status} (${res.body && res.body.slice(0, 200)})`,
    );
  }
  const body = res.json();
  const simple = (body.items || []).filter(
    (p) => p.type === "simple" && p.priceCents > 0,
  );
  if (simple.length === 0) {
    throw new Error(
      "setup: no `simple` products in catalog — seed staging with products first",
    );
  }
  console.log(`setup: ${simple.length} simple products available for ordering`);
  return { products: simple.map((p) => ({ id: p.id, price: p.priceCents })) };
}

// ---------------------------------------------------------------------------
// Per-VU helpers
// ---------------------------------------------------------------------------

function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function bearer(token) {
  return {
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  };
}

function postJSON(url, token, body, extraHeaders) {
  const opts = bearer(token);
  if (extraHeaders) Object.assign(opts.headers, extraHeaders);
  return http.post(url, JSON.stringify(body), opts);
}

function buildCartItems(products) {
  const n = randomIntBetween(1, 3);
  const picked = [];
  for (let i = 0; i < n; i++) {
    picked.push({
      productId: pick(products).id,
      quantity: randomIntBetween(1, 2),
    });
  }
  return picked;
}

// ---------------------------------------------------------------------------
// Scenarios
// ---------------------------------------------------------------------------

export function customer(data) {
  const sessionToken = pick(tokens.customer_sessions);

  group("customer/list-products", () => {
    const r = http.get(`${BASE_URL}/v1/products`, bearer(sessionToken));
    check(r, { "products 200": (res) => res.status === 200 });
  });

  let orderId;
  group("customer/create-order", () => {
    const r = postJSON(
      `${BASE_URL}/v1/orders`,
      sessionToken,
      {
        items: buildCartItems(data.products),
        contactEmail: `loadtest+vu${__VU}@loadtest.local`,
      },
      { "Idempotency-Key": uuidv4() },
    );
    const ok = check(r, {
      "create 2xx": (res) => res.status === 200 || res.status === 201,
    });
    if (ok) orderId = r.json("id");
  });

  if (!orderId) return;

  group("customer/pay-twint", () => {
    const r = postJSON(
      `${BASE_URL}/v1/orders/${orderId}/payment`,
      sessionToken,
      { method: "twint", returnUrl: "https://loadtest.local/return" },
      { "Idempotency-Key": uuidv4() },
    );
    check(r, { "pay 2xx": (res) => res.status === 200 || res.status === 201 });
  });

  group("customer/poll-payment", () => {
    const r = http.get(
      `${BASE_URL}/v1/orders/${orderId}/payment`,
      bearer(sessionToken),
    );
    check(r, { "poll 2xx": (res) => res.status === 200 });
  });

  sleep(randomIntBetween(1, 3));
}

export function admin(data) {
  const sessionToken = pick(tokens.admin_sessions);
  const today = new Date();
  const from = new Date(today);
  from.setHours(0, 0, 0, 0);
  const to = new Date(today);
  to.setHours(24, 0, 0, 0);
  const qs = `date_from=${encodeURIComponent(from.toISOString())}&date_to=${encodeURIComponent(to.toISOString())}`;

  group("admin/dashboard-fetch", () => {
    // Mirrors the parallel fetch in bfs-web-app/app/admin/page.tsx.
    const responses = http.batch([
      ["GET", `${BASE_URL}/v1/orders?${qs}`, null, bearer(sessionToken)],
      ["GET", `${BASE_URL}/v1/products`, null, bearer(sessionToken)],
    ]);
    check(responses[0], { "orders 200": (r) => r.status === 200 });
    check(responses[1], { "products 200": (r) => r.status === 200 });
  });

  sleep(randomIntBetween(3, 8));
}

export function station(data) {
  const posToken = pick(tokens.pos_tokens);
  const stationToken = pick(tokens.station_tokens);

  // POS creates and pays for an order (gratis_guest → no payment integration).
  let orderId;
  group("station/pos-create-order", () => {
    const r = postJSON(
      `${BASE_URL}/v1/orders`,
      posToken,
      { items: buildCartItems(data.products) },
      { "Idempotency-Key": uuidv4() },
    );
    if (
      check(r, {
        "pos create 2xx": (res) => res.status === 200 || res.status === 201,
      })
    ) {
      orderId = r.json("id");
    }
  });

  if (!orderId) return;

  group("station/pos-pay-cash", () => {
    const r = postJSON(
      `${BASE_URL}/v1/orders/${orderId}/payment`,
      posToken,
      { method: "cash" },
      { "Idempotency-Key": uuidv4() },
    );
    check(r, {
      "pos pay 2xx": (res) => res.status === 200 || res.status === 201,
    });
  });

  // Station scans the order and redeems items assigned to it.
  group("station/redeem", () => {
    const r = postJSON(
      `${BASE_URL}/v1/stations/redeem`,
      stationToken,
      { orderId },
      { "Idempotency-Key": uuidv4() },
    );
    check(r, {
      "redeem 2xx": (res) => res.status === 200 || res.status === 201,
    });
  });
}

// k6 load test for the bless2n-food-system backend.
//
// Env: BASE_URL, PROFILE = smoke | baseline | stress
//
// Requires the target env to have: tokens seeded via seed.sql, an approved
// STATION device with assigned products, and Payrexx unconfigured (so TWINT
// falls through to MarkOrderPaidDev). See loadtest/README.md.

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

// `baseline` mirrors the production dinner-rush shape (18:30–22:30), compressed
// to ~12 minutes. Multiply stage durations to simulate a full evening.
const profiles = {
  smoke: {
    customer: {
      executor: "constant-vus",
      exec: "customer",
      vus: 1,
      duration: "30s",
      tags: { scenario: "customer" },
    },
    admin: {
      executor: "constant-vus",
      exec: "admin",
      vus: 1,
      duration: "30s",
      tags: { scenario: "admin" },
    },
    station: {
      executor: "constant-vus",
      exec: "station",
      vus: 1,
      duration: "30s",
      tags: { scenario: "station" },
    },
  },

  baseline: {
    customer: {
      executor: "ramping-arrival-rate",
      exec: "customer",
      startRate: 1,
      timeUnit: "1s",
      preAllocatedVUs: 30,
      maxVUs: 80,
      stages: [
        { target: 2, duration: "1m" }, // warmup — avoids measuring cold-start
        { target: 7, duration: "2m" },
        { target: 7, duration: "5m" }, // sustained dinner peak
        { target: 3, duration: "2m" },
        { target: 0, duration: "1m" },
      ],
      tags: { scenario: "customer" },
    },
    admin: {
      executor: "constant-arrival-rate",
      exec: "admin",
      rate: 1,
      timeUnit: "10s",
      preAllocatedVUs: 2,
      maxVUs: 5,
      duration: "11m",
      tags: { scenario: "admin" },
    },
    station: {
      executor: "ramping-arrival-rate",
      exec: "station",
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
      exec: "customer",
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
  maxRedirects: 0,
};

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

  // Cash instead of TWINT — avoids Payrexx sandbox flakiness masking backend perf.
  group("customer/pay-cash", () => {
    const r = postJSON(
      `${BASE_URL}/v1/orders/${orderId}/payment`,
      sessionToken,
      { method: "cash" },
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

  // Mirrors the parallel fetch in bfs-web-app/app/admin/page.tsx.
  group("admin/dashboard-fetch", () => {
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

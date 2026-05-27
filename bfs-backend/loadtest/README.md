# Load testing

k6-based load tests for the backend API. Mirrors the four production traffic
shapes — customer ordering, admin dashboard, POS, and station redeem — at
weights and durations roughly matching the 18:30–22:30 dinner-rush window from
production latency graphs.

## What you need

- **k6** (`brew install k6` or [k6.io/docs/get-started/installation](https://k6.io/docs/get-started/installation/))
- **psql** with access to the target environment's `DATABASE_URL`
- A staging deploy that you are sure is not production
- One approved STATION device in that environment with products assigned (via
  `/admin/stations`). Without this the redeem scenario will return empty
  matches and report low latency for the wrong reason.

## One-time setup against staging

```bash
# 1. Point at staging
export DATABASE_URL='postgres://…neondb…'           # staging Neon URL
export BASE_URL='https://api.staging.example.com'   # staging API URL

# 2. Find your staging station device ID (paste into next command)
psql "$DATABASE_URL" -c "SELECT id, name FROM device WHERE type='STATION' AND status='approved';"

# 3. Seed loadtest users, sessions, and device bindings
psql "$DATABASE_URL" \
  -v station_device_id=<uuid-from-step-2> \
  -v pos_device_id=NULL \
  -f loadtest/seed.sql
```

The seed creates:

- 20 customer users + sessions (tokens: `loadtest_customer_session_001…020`)
- 3 admin users + sessions (`loadtest_admin_session_001…003`)
- 3 STATION device bindings (`loadtest_station_token_001…003`)
- 2 POS device bindings (`loadtest_pos_token_001…002`)

All rows are tagged with `loadtest_` prefixes or `loadtest+` emails for easy
cleanup. The seed is idempotent — re-running refreshes session expiry without
duplicating rows.

## Running

```bash
# Quick sanity check (1 VU per scenario, 30s)
just loadtest BASE_URL="$BASE_URL" PROFILE=smoke

# Dinner-rush simulation (~12 minutes, ramps to ~10 RPS combined)
just loadtest BASE_URL="$BASE_URL" PROFILE=baseline

# Stress test (ramps customer scenario to 50 RPS until failures)
just loadtest BASE_URL="$BASE_URL" PROFILE=stress
```

Or invoke k6 directly:

```bash
k6 run -e BASE_URL="$BASE_URL" -e PROFILE=baseline loadtest/k6.js
```

## Browser test — end-to-end with real Chromium

`browser.js` drives the full Next.js → backend customer order flow inside real
Chromium pages and captures **Core Web Vitals** (LCP, FCP, INP, CLS, TTFB) per
navigation. Use it alongside the API test to answer "how does the user
_experience_ the system under load," not just "how fast does the API answer."

```bash
# Smoke: 1 VU, 3 iterations — verify the script can drive the flow
WEB_URL="https://staging.example.com" PROFILE=smoke just loadtest-browser

# Baseline: 8 concurrent browsers, 10 min
WEB_URL="https://staging.example.com" PROFILE=baseline just loadtest-browser

# Stress: ramp to 12 browsers
WEB_URL="https://staging.example.com" PROFILE=stress just loadtest-browser

# Full-stack: API + browser in parallel — the realistic combined-load picture
BASE_URL="https://api.staging.example.com" \
WEB_URL="https://staging.example.com" \
PROFILE=baseline \
just loadtest-full
```

`WEB_URL` is the **public Next.js host**, not the backend API host (`BASE_URL`).

The browser flow opens `/food` (the menu), adds two simple products to cart,
navigates to `/food/checkout`, clicks "Mit TWINT bezahlen", and waits briefly
before the next iteration. Mobile viewport (iPhone-sized) by default since most
real customers are on phones.

### Resource cost

Each browser VU runs one Chromium tab — ~150-250 MB resident. On the 24 GB M4
Pro the `baseline` profile (8 VUs) sits at ~2 GB, the `stress` profile (12 VUs)
at ~3 GB. Sustained higher concurrency requires monitoring Activity Monitor;
swapping kills your timing measurements.

### Web Vitals thresholds

The browser test enforces these targets (tweak in `browser.js`):

| Metric | p75    | p95    | Notes                                       |
| ------ | ------ | ------ | ------------------------------------------- |
| LCP    | 2.5 s  | 4 s    | Google "good" threshold                     |
| FCP    | 1.8 s  | 3 s    | Google "good" threshold                     |
| INP    | 200 ms | 500 ms | Interaction-to-next-paint (post-March 2024) |
| CLS    | 0.1    | —      | Layout shift                                |

## What the scenarios do

### `k6.js` (API)

| Scenario | Auth                 | Flow                                                             | Endpoints exercised                                            |
| -------- | -------------------- | ---------------------------------------------------------------- | -------------------------------------------------------------- |
| customer | session              | list products → create order → TWINT pay → poll status           | `/v1/products`, `/v1/orders`, `/v1/orders/{id}/payment`        |
| admin    | admin session        | parallel fetch (orders for today + products), like the dashboard | `/v1/orders?date_from=…`, `/v1/products`                       |
| station  | POS + station device | POS creates+pays an order, station redeems it                    | `/v1/orders`, `/v1/orders/{id}/payment`, `/v1/stations/redeem` |

### `browser.js` (UX)

| Scenario | Viewport              | Flow                                                     |
| -------- | --------------------- | -------------------------------------------------------- |
| customer | iPhone-sized (mobile) | open /food → add 2 products → /food/checkout → click pay |

The TWINT path hits `MarkOrderPaidDev` (no real Payrexx call) **only if Payrexx
is not configured in the target env**. Confirm with the operator that staging
has no `PAYREXX_INSTANCE` / `PAYREXX_API_SECRET` set, or you will generate real
gateway requests.

## Reading the results

k6 prints a summary at the end. The thresholds in `k6.js` will fail the run if:

- `http_req_failed` rate exceeds 2%
- customer p95 exceeds 1500 ms
- station p95 exceeds 1000 ms
- admin p95 exceeds 2000 ms

Use the per-scenario tagged metrics (`http_req_duration{scenario:customer}`) to
compare against the production latency graph that motivated the optimizations.

## Cleanup

```bash
psql "$DATABASE_URL" -f loadtest/cleanup.sql
```

Removes all `loadtest_*` users, sessions, device bindings, and any orders with
`loadtest+…@loadtest.local` contact emails.

## Tuning

- `tokens.json` — increase token pool size if you saturate sessions (k6 will
  start hitting the same session repeatedly under high VU counts). For the
  in-memory auth cache this is fine; for the underlying DB lookup behaviour you
  want broader coverage.
- `profiles.baseline.stages` in `k6.js` — multiply durations to simulate a full
  evening instead of a compressed 12-minute slice.
- `thresholds` — tighten as you ship optimizations and want to ratchet up the
  bar.

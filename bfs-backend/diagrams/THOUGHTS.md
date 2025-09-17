# Bless2n Food System — Database Notes (ERD, Requirements & Decisions)

*Last updated: 2025-09-17*

## 1) Purpose & Scope

This document captures the shared understanding for the database model that supports:

* **Simple products** (e.g., Fries, Red Bull, Cheeseburger)
* **Menu products** (e.g., “Menu Small”, “Menu Big”) defined as curated combinations of simple products
* **Ordering** with exact choices stored (e.g., which burger/drink)
* **Redemption** at temporary **stations** (pickup-only)
* **Global inventory** movements (not station-scoped)

Focus is on the **ERD** and the behavioral rules around menus, orders, redemption, and inventory.

---

## 2) Core Requirements (Agreed)

* **Two product types:** `simple` and `menu`, both live in the same `products` collection.
* **Menus are products** with their own price, image, category, active flag, etc.
* **Menu composition:** via `menu_slots` (e.g., *Fries*, *Burger*, *Drink*) and curated **allowed options** via `menu_slot_items` (each option is a `simple` product).
* **Admin curation:** Add/remove allowed simple products per slot in an admin UI (no `menu_slot_items.is_active`; deactivation is done on the product itself or by removing the slot item).
* **Station usage:** Stations are **pickup-only**, ephemeral per event/day (\~24h). No per-station price or stock. Stations list which **simple** products can be redeemed there (`station_products`).
* **Ordering model (parent/child):**

  * One **parent** `order_item` for the `menu` carrying the **menu price**.
  * One or more **child** `order_item`s for the selected **simple** items (price 0), each linked to the parent and **stamped with `menu_slot_id`** (plus `menu_slot_name` snapshot).
* **Redemption:** All-or-nothing per `order_item` (`is_redeemed` is a **boolean**; no partial redemption).
* **Inventory:** Global (not station-scoped). Stock movements are recorded in an **inventory ledger** on the **child simple items** (sales negative; refunds/restock positive).
* **Money:** Use **integer cents** consistently (e.g., `price_per_unit_cents`, `line_total_cents`, `orders.total_cents`) to avoid rounding.
* **Categories:** Keep as-is (no split for simple/menu).
* **Admin invites:** Include `status` (e.g., `pending|accepted|revoked|expired`) and `used_at`.

---

## 3) ERD (Narrative Overview)

### Users & Auth

* **users** — accounts with roles `admin|customer`, verification & disable flags.
* **admin\_invites** — invited\_by → users, with `status`, `expires_at`, `used_at`.
* **otp\_tokens** — short-lived login/password reset tokens.
* **refresh\_tokens** — hashed, revocable, token families.

### Operations

* **stations** — approval workflow; pickup-only; ephemeral per event/day.
* **devices** — bound to a station; can redeem items.
* **device\_requests** — station device access approvals.
* **station\_products** — which **simple** products can be redeemed at a given station (no prices, no stock).

### Catalog

* **categories** — assigns a category to any product.
* **products** — `simple` or `menu`; own price, image, active.
* **menu\_slots** — for menu products only (e.g., *Fries*, *Burger*, *Drink*), ordered by `sequence`.
* **menu\_slot\_items** — allowed **simple** products for a slot (admin curated). No surcharge and no slot-level active flag.

### Ordering & Inventory

* **orders** — customer (optional), status (`pending|paid|cancelled|refunded`), monetary totals in cents.
* **order\_items**

  * **Parent**: `product_id` points to a `menu`, carries the **menu price**.
  * **Child**: `product_id` points to a **simple** product, `parent_item_id` set, `menu_slot_id` set, **price 0** (unless a future rule changes this).
  * Snapshots: `title` and `menu_slot_name` retained for durability.
  * Redemption: `is_redeemed` boolean + `redeemed_at`, `redeemed_station_id`, `redeemed_device_id`.
* **inventory\_ledger** — global stock movements for **simple** products only (negative on sale, positive on refund/restock/manual/correction).

---

## 4) Behavioral Rules & Invariants

### Menu composition & validation

* A `menu` product **must** have ≥1 `menu_slots`.
* Each `menu_slot` **must** whitelist allowed **simple** products via `menu_slot_items`.
* On order placement, for each chosen child item:

  * The `product_id` **must** be a `simple` product.
  * The `product_id` **must** be allowed for the referenced `menu_slot_id`.
  * The `parent_item_id` **must** point to a parent `order_item` with a `menu` product.

### Station & redemption

* Redemption **only** at stations that list the **simple** product in `station_products`.
* `is_redeemed` flips to `true` once scanned/picked up (no partial per item).
* A menu’s **parent** line is **not** directly redeemable; redemption happens on child simple lines (or you mark parent as redeemed when all children are redeemed—implementation choice, but boolean is all-or-nothing).

### Inventory

* Decrement inventory **only for child simple lines** on sale (or on payment confirmation if you choose to “reserve” stock earlier).
* Refunds add positive deltas for the same simple products.
* Inventory is **global** (no station dimension).

### Money

* All monetary values in **cents** integers.
* `orders.total_cents` equals the sum of line totals (parent menu price + any priced simple lines—currently zero).

### Snapshots & durability

* `order_items.title` and `menu_slot_name` are **snapshots** to preserve order readability despite catalog changes.
* Product deactivation after order placement **must not** break redemption; snapshot + historical references ensure auditability.

---

## 5) Typical Flows (Condensed)

**Menu order (“Menu Big”):**

1. Create **parent** `order_item` with `product_id = Menu Big`, `price_per_unit_cents = menu price`.
2. For each slot:

   * Create **child** `order_item` with `product_id = chosen simple`, `parent_item_id = parent`, `menu_slot_id` set, `price_per_unit_cents = 0`.
3. On payment: set `orders.status = paid`; record **inventory\_ledger** deltas (negative) for **child simple** items.

**Redemption at station:**

1. Device scans an `order_item` child (simple product).
2. Validate:

   * Order is `paid`, not cancelled/refunded.
   * Station allows redemption for that simple product.
3. Set `is_redeemed = true`, stamp `redeemed_at`, station & device refs.
4. Optionally cascade mark the parent menu as redeemed when all children are redeemed (your boolean model can store this at parent level too if desired).

**Refund (full order):**

1. Set `orders.status = refunded`.
2. Create **inventory\_ledger** positive deltas equal to the previously sold child simple items.
3. Mark relevant order\_items as reverted (implementation detail; not required by schema but good for audit).

---

## 6) Admin & Operations Notes

* **Admin curation:** To remove an option from a menu, **delete the `menu_slot_item`** or deactivate the **simple product**. No slot-level `is_active`.
* **Stations:** Created/approved per event; can be cleaned up after the event. They never set prices or stock.
* **Analytics:**

  * Menus sold = count **parent** menu lines.
  * Components used = count **child** simple lines (per slot and per product).
  * Revenue attribution stays on the parent; component attribution can be derived analytically.
* **Security:** Devices belong to stations; device requests are approved by admins. All redemption actions must be authenticated/authorized.

---

## 8) Example: “Menu Big” (Conceptual)

* **Menu Big** (`products.type = menu`, `price_cents = 1500`)

  * Slots:

    * *Fries* → allowed: Fries
    * *Burger* → allowed: Beef Burger, Vegan Burger
    * *Drink* → allowed: curated list of drinks (e.g., Coke, Red Bull)
* **Order Items:**

  * Parent: `Menu Big` @ 1500
  * Child 1: `Fries` (slot: Fries) @ 0
  * Child 2: `Beef Burger` (slot: Burger) @ 0
  * Child 3: `Coke` (slot: Drink) @ 0
* **Inventory movements:** −1 Fries, −1 Beef Burger, −1 Coke.
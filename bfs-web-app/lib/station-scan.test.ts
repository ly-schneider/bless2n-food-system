import { describe, expect, it } from "vitest"
import { parseScan } from "./station-scan"

const LEGACY_UUID = "0197c7a4-1111-7000-8000-000000000001"

describe("parseScan", () => {
  it("parses an order pickup URL", () => {
    expect(parseScan("https://order.bless2n.ch/o/HS79U1yH7Zd3")).toEqual({
      kind: "order",
      id: "HS79U1yH7Zd3",
    })
  })

  it("parses an order pickup URL with trailing slash", () => {
    expect(parseScan("http://localhost:3000/o/HS79U1yH7Zd3/")).toEqual({
      kind: "order",
      id: "HS79U1yH7Zd3",
    })
  })

  it("parses a bare order id", () => {
    expect(parseScan("  HS79U1yH7Zd3 ")).toEqual({ kind: "order", id: "HS79U1yH7Zd3" })
  })

  it("parses a campaign payload", () => {
    expect(parseScan("CAMP:tkn_camp___1")).toEqual({ kind: "campaign", id: "tkn_camp___1" })
  })

  it("is case-insensitive on the campaign prefix only", () => {
    expect(parseScan("camp:tkn_camp___1")).toEqual({ kind: "campaign", id: "tkn_camp___1" })
  })

  it("parses a legacy uuid order id (bare)", () => {
    expect(parseScan(`  ${LEGACY_UUID} `)).toEqual({ kind: "order", id: LEGACY_UUID })
  })

  it("parses a legacy uuid order id via pickup URL", () => {
    expect(parseScan(`https://order.bless2n.ch/o/${LEGACY_UUID}`)).toEqual({
      kind: "order",
      id: LEGACY_UUID,
    })
  })

  it("parses a legacy uuid campaign token", () => {
    expect(parseScan(`CAMP:${LEGACY_UUID}`)).toEqual({ kind: "campaign", id: LEGACY_UUID })
  })

  it("rejects campaign payloads with invalid tokens", () => {
    expect(parseScan("CAMP:not-an-id")).toBeNull()
  })

  it("rejects ids with confusable characters (0, O, I, l)", () => {
    expect(parseScan("O0Il0OIl0OIl")).toBeNull()
  })

  it("does not match a 12-char window inside a hostname", () => {
    expect(parseScan("https://very-long-tenant.bless2n.ch/")).toBeNull()
  })

  it("rejects empty and garbage input", () => {
    expect(parseScan("")).toBeNull()
    expect(parseScan("   ")).toBeNull()
    expect(parseScan("hello world")).toBeNull()
  })
})

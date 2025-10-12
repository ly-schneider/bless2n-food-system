import { describe, expect, it } from "vitest"
import { cn, formatChf } from "./utils"

describe("utils", () => {
  it("cn merges class names", () => {
    expect(cn("a", false && "b", "c")).toBe("a c")
  })
  it("formatChf formats cents to CHF", () => {
    expect(formatChf(0)).toContain("CHF")
    expect(formatChf(1990)).toMatch(/19/) // 19.90
  })
})

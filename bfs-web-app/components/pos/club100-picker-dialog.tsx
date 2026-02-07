"use client"

import { Search } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { type Club100Person, listClub100People } from "@/lib/api/club100"
import { formatChf } from "@/lib/utils"
import type { CartItem } from "@/types/cart"
import type { Club100Discount, Club100DiscountItem } from "@/types/order-queue"

interface Club100PickerDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSelect: (person: Club100Person, discount: Club100Discount | null) => void
  token: string
  cartItems: CartItem[]
  freeProductIds: string[]
}

export function Club100PickerDialog({
  open,
  onOpenChange,
  onSelect,
  token,
  cartItems,
  freeProductIds,
}: Club100PickerDialogProps) {
  const [people, setPeople] = useState<Club100Person[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState("")

  const loadPeople = useCallback(async () => {
    if (!token) return
    setLoading(true)
    setError(null)
    try {
      const result = await listClub100People(token)
      setPeople(result)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Fehler beim Laden")
    } finally {
      setLoading(false)
    }
  }, [token])

  useEffect(() => {
    if (open) {
      loadPeople()
      setSearch("")
    }
  }, [open, loadPeople])

  const filtered = useMemo(() => {
    if (!search.trim()) return people
    const q = search.toLowerCase()
    return people.filter(
      (p) => p.firstName.toLowerCase().includes(q) || p.lastName.toLowerCase().includes(q)
    )
  }, [people, search])

  const calculateDiscount = useCallback(
    (person: Club100Person): Club100Discount | null => {
      if (freeProductIds.length === 0 || person.remaining <= 0) {
        return null
      }

      const discountedItems: Club100DiscountItem[] = []
      let remainingRedemptions = person.remaining
      let totalDiscountCents = 0

      for (const item of cartItems) {
        if (remainingRedemptions <= 0) break

        const isEligible = freeProductIds.includes(item.product.id)
        if (!isEligible) continue

        const discountedQuantity = Math.min(item.quantity, remainingRedemptions)
        remainingRedemptions -= discountedQuantity
        totalDiscountCents += discountedQuantity * item.totalPriceCents

        discountedItems.push({
          cartItemId: item.id,
          productId: item.product.id,
          productName: item.product.name,
          quantity: item.quantity,
          discountedQuantity,
          unitPriceCents: item.totalPriceCents,
        })
      }

      if (discountedItems.length === 0) {
        return null
      }

      return {
        person: {
          id: person.id,
          firstName: person.firstName,
          lastName: person.lastName,
          remaining: person.remaining,
          max: person.max,
        },
        discountedItems,
        totalDiscountCents,
        freeProductIds,
      }
    },
    [cartItems, freeProductIds]
  )

  const getPersonPreview = useCallback(
    (person: Club100Person) => {
      const discount = calculateDiscount(person)
      if (!discount) {
        return { eligibleCount: 0, discountCents: 0 }
      }
      const eligibleCount = discount.discountedItems.reduce((sum, d) => sum + d.discountedQuantity, 0)
      return { eligibleCount, discountCents: discount.totalDiscountCents }
    },
    [calculateDiscount]
  )

  const handleSelect = useCallback(
    (person: Club100Person) => {
      const discount = calculateDiscount(person)
      onSelect(person, discount)
    },
    [calculateDiscount, onSelect]
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>100 Club Mitglied wählen</DialogTitle>
        </DialogHeader>

        <div className="relative">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Nach Name suchen..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        <div className="flex-1 min-h-0 overflow-y-auto space-y-2 py-2">
          {loading && <div className="text-center text-muted-foreground py-4">Laden...</div>}
          {error && <div className="text-center text-red-600 py-4">{error}</div>}
          {!loading && !error && filtered.length === 0 && (
            <div className="text-center text-muted-foreground py-4">Keine Mitglieder gefunden</div>
          )}
          {!loading &&
            !error &&
            filtered.map((person) => {
              const noRedemptionsLeft = person.remaining <= 0
              const preview = getPersonPreview(person)
              const noEligibleProducts = preview.eligibleCount === 0 && !noRedemptionsLeft

              return (
                <button
                  key={person.id}
                  disabled={noRedemptionsLeft}
                  onClick={() => handleSelect(person)}
                  className={`w-full flex items-center justify-between rounded-lg border px-4 py-3 text-left transition-colors ${
                    noRedemptionsLeft
                      ? "opacity-50 cursor-not-allowed bg-muted"
                      : "hover:bg-accent cursor-pointer"
                  }`}
                >
                  <div className="flex-1">
                    <div className="font-medium">
                      {person.firstName} {person.lastName}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {person.remaining} von {person.max} verbleibend
                    </div>
                    {!noRedemptionsLeft && preview.eligibleCount > 0 && (
                      <div className="text-sm text-green-600 font-medium mt-1">
                        {preview.eligibleCount} Produkt{preview.eligibleCount > 1 ? "e" : ""} rabattfähig
                        ({formatChf(preview.discountCents)} Rabatt)
                      </div>
                    )}
                    {noEligibleProducts && (
                      <div className="text-sm text-amber-600 mt-1">
                        Keine rabattfähigen Produkte im Warenkorb
                      </div>
                    )}
                  </div>
                  {!noRedemptionsLeft && (
                    <Button size="sm" variant={noEligibleProducts ? "outline" : "default"}>
                      Wählen
                    </Button>
                  )}
                </button>
              )
            })}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Abbrechen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

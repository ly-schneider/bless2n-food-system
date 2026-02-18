"use client"

import { Crown, Handshake, Star, Users } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"

type GratisType = "guest" | "vip" | "staff" | "100club"

interface GratisTypeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSelect: (type: GratisType) => void
}

export function GratisTypeDialog({ open, onOpenChange, onSelect }: GratisTypeDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Gratis-Typ wählen</DialogTitle>
        </DialogHeader>
        <div className="grid grid-cols-2 gap-3">
          <Button
            className="flex h-28 flex-col items-center justify-center gap-2 rounded-xl"
            variant="outline"
            onClick={() => onSelect("guest")}
            aria-label="Gäste"
          >
            <Users className="size-10" />
            <span className="text-base font-medium">Gäste</span>
          </Button>
          <Button
            className="flex h-28 flex-col items-center justify-center gap-2 rounded-xl"
            variant="outline"
            onClick={() => onSelect("vip")}
            aria-label="VIP"
          >
            <Star className="size-10" />
            <span className="text-base font-medium">VIP</span>
          </Button>
          <Button
            className="flex h-28 flex-col items-center justify-center gap-2 rounded-xl"
            variant="outline"
            onClick={() => onSelect("staff")}
            aria-label="Mitarbeiter"
          >
            <Handshake className="size-10" />
            <span className="text-base font-medium">Mitarbeiter</span>
          </Button>
          <Button
            className="flex h-28 flex-col items-center justify-center gap-2 rounded-xl"
            variant="outline"
            onClick={() => onSelect("100club")}
            aria-label="100 Club"
          >
            <Crown className="size-10" />
            <span className="text-base font-medium">100 Club</span>
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

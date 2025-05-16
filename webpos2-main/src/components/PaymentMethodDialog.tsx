import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { CreditCard, Banknote, QrCode, UserCircle2, Star, Gift, Heart, Coffee, ArrowLeft } from "lucide-react"
import { useState } from "react"

interface PaymentMethodDialogProps {
  isOpen: boolean;
  onSelect: (method: "CSH" | "CRE" | "TWI" | "EMP" | "VIP" | "KUL" | "GUT" | "DIV") => void;
  onClose: () => void;
}

export function PaymentMethodDialog({
  isOpen,
  onSelect,
  onClose,
}: PaymentMethodDialogProps) {
  const [showSpecial, setShowSpecial] = useState(false);

  const handleSpecialBack = () => {
    setShowSpecial(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[600px] w-full">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {showSpecial && (
              <Button variant="ghost" size="icon" onClick={handleSpecialBack} className="h-6 w-6">
                <ArrowLeft className="h-4 w-4" />
              </Button>
            )}
            Zahlungsmethode wählen
          </DialogTitle>
        </DialogHeader>
        {!showSpecial ? (
          <>
            <div className="grid grid-cols-2 gap-6 p-4">
              <Button
                variant="outline"
                className={`flex flex-col items-center justify-center h-40 ${isOpen ? "border-primary" : ""}`}
                onClick={() => onSelect("CSH")}
              >
                <Banknote className="h-12 w-12 mb-3" />
                <span className="text-lg">Bar</span>
              </Button>
              <Button
                variant="outline"
                className={`flex flex-col items-center justify-center h-40 ${isOpen ? "border-primary" : ""}`}
                onClick={() => onSelect("CRE")}
              >
                <CreditCard className="h-12 w-12 mb-3" />
                <span className="text-lg">Karte</span>
              </Button>
              <Button
                variant="outline"
                className={`flex flex-col items-center justify-center h-40 ${isOpen ? "border-primary" : ""}`}
                onClick={() => onSelect("TWI")}
              >
                <QrCode className="h-12 w-12 mb-3" />
                <span className="text-lg">Twint</span>
              </Button>
              <Button
                variant="outline"
                className="flex flex-col items-center justify-center h-40 opacity-60 hover:opacity-80"
                onClick={() => onSelect("EMP")}
              >
                <UserCircle2 className="h-12 w-12 mb-3" />
                <span className="text-lg">Mitarbeiter</span>
              </Button>
            </div>
            <div className="px-4 pb-4">
              <Button
                variant="outline"
                className="w-full h-12 opacity-60 hover:opacity-80"
                onClick={() => setShowSpecial(true)}
              >
                <Star className="h-5 w-5 mr-2" />
                <span className="text-lg">Spezial</span>
              </Button>
            </div>
          </>
        ) : (
          <div className="grid grid-cols-2 gap-6 p-4">
            <Button
              variant="outline"
              className="flex flex-col items-center justify-center h-40"
              onClick={() => onSelect("VIP")}
            >
              <Star className="h-12 w-12 mb-3" />
              <span className="text-lg">VIP Gast</span>
            </Button>
            <Button
              variant="outline"
              className="flex flex-col items-center justify-center h-40"
              onClick={() => onSelect("KUL")}
            >
              <Heart className="h-12 w-12 mb-3" />
              <span className="text-lg">Kullanz</span>
            </Button>
            <Button
              variant="outline"
              className="flex flex-col items-center justify-center h-40"
              onClick={() => onSelect("GUT")}
            >
              <Gift className="h-12 w-12 mb-3" />
              <span className="text-lg">Gutschein</span>
            </Button>
            <Button
              variant="outline"
              className="flex flex-col items-center justify-center h-40"
              onClick={() => onSelect("DIV")}
            >
              <Coffee className="h-12 w-12 mb-3" />
              <span className="text-lg">Anderes</span>
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

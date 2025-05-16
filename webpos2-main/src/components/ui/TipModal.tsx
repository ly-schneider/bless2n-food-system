import { FC, useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from './button'
import { Input } from './input'

interface TipModalProps {
  currentTotal: number
  minimumTotal: number
  onClose: () => void
  onSave: (newTotal: number) => void
}

export const TipModal: FC<TipModalProps> = ({ 
  currentTotal, 
  minimumTotal, 
  onClose, 
  onSave 
}) => {
  const [amount, setAmount] = useState(currentTotal)
  const [error, setError] = useState<string>('')

  useEffect(() => {
    setAmount(Math.max(currentTotal, minimumTotal))
  }, [currentTotal, minimumTotal])

  const handleAmountChange = (value: string) => {
    const parsedValue = parseFloat(value)
    
    if (value === '') {
      setAmount(minimumTotal)
      setError('')
      return
    }

    if (isNaN(parsedValue)) {
      setError('Bitte geben Sie einen gültigen Betrag ein')
      return
    }

    if (parsedValue > 10000) {
      setError('Der maximale Betrag ist CHF 10\'000')
      return
    }

    if (parsedValue < 0) {
      setError('Der Betrag kann nicht negativ sein')
      return
    }

    setAmount(parsedValue)
  }

  useEffect(() => {
    if (amount < minimumTotal) {
      setError(`Betrag muss mindestens CHF ${minimumTotal.toFixed(2)} sein`)
    } else if (amount > 10000) {
      setError('Der maximale Betrag ist CHF 10\'000')
    } else {
      setError('')
    }
  }, [amount, minimumTotal])

  const handleSave = () => {
    if (amount >= minimumTotal && amount <= 10000 && !error) {
      onSave(amount)
    }
  }

  return (
    <Dialog open={true} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="text-2xl">Gesamtbetrag eingeben</DialogTitle>
        </DialogHeader>
        <div className="py-6 space-y-4">
          <div className="relative">
            <DialogDescription className="text-lg">
              Gesamtbetrag eingeben (mindestens CHF {minimumTotal.toFixed(2)}):
            </DialogDescription>
            <Input 
              type="number" 
              inputMode="decimal"
              pattern="[0-9]*"
              value={amount} 
              onChange={(e) => handleAmountChange(e.target.value)} 
              className="mt-4 w-full text-3xl h-16 px-4"
              min={minimumTotal}
              max={10000}
              step="0.05"
            />
            {error && (
              <div className="absolute left-0 right-0 -top-12 bg-destructive/95 text-destructive-foreground px-4 py-2 rounded-md shadow-lg">
                <p className="text-base">{error}</p>
              </div>
            )}
            <div className="grid grid-cols-3 gap-4 mt-6">
              {[1, 2, 3, 4, 5, 6, 7, 8, 9, '.', 0, '<'].map((num) => (
                <Button
                  key={num}
                  variant="outline"
                  className="h-20 text-2xl font-semibold hover:bg-primary hover:text-primary-foreground"
                  onClick={() => {
                    if (num === '<') {
                      handleAmountChange(amount.toString().slice(0, -1) || '0');
                    } else {
                      const newAmount = amount === minimumTotal ? num.toString() : amount.toString() + num;
                      handleAmountChange(newAmount);
                    }
                  }}
                >
                  {num === '<' ? '⌫' : num}
                </Button>
              ))}
            </div>
          </div>
        </div>
        <div className="flex justify-end space-x-4">
          <Button variant="outline" onClick={onClose} className="text-lg px-6 py-3">Abbrechen</Button>
          <Button 
            onClick={handleSave} 
            disabled={!!error || amount < minimumTotal || amount > 10000}
            className="text-lg px-6 py-3"
          >
            Speichern
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

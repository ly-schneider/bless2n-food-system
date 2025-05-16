import { FC, useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from './button'
import { Input } from './input'

interface RoundUpModalProps {
  currentAmount: number
  baseAmount: number
  onClose: () => void
  onSave: (newAmount: number) => void
}

export const RoundUpModal: FC<RoundUpModalProps> = ({ currentAmount, baseAmount, onClose, onSave }) => {
  const [amount, setAmount] = useState(currentAmount)
  const [error, setError] = useState<string>('')

  useEffect(() => {
    if (amount < baseAmount) {
      setError(`Amount must be at least CHF ${baseAmount.toFixed(2)}`)
    } else {
      setError('')
    }
  }, [amount, baseAmount])

  const handleSave = () => {
    if (amount >= baseAmount) {
      onSave(amount)
    }
  }

  return (
    <Dialog open={true} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Edit Round-Up Amount</DialogTitle>
        </DialogHeader>
        <div className="py-4">
          <DialogDescription>
            Enter the total amount (must be at least CHF {baseAmount.toFixed(2)}):
          </DialogDescription>
          <Input 
            type="number" 
            value={amount} 
            onChange={(e) => setAmount(Math.max(0, parseFloat(e.target.value) || baseAmount))} 
            className="mt-2"
            min={baseAmount}
            step="0.05"
          />
          {error && (
            <p className="text-sm text-red-500 mt-1">{error}</p>
          )}
          {amount > baseAmount && (
            <p className="text-sm text-gray-600 mt-1">
              Donation amount: CHF {(amount - baseAmount).toFixed(2)}
            </p>
          )}
        </div>
        <div className="flex justify-end space-x-2">
          <Button onClick={onClose}>Cancel</Button>
          <Button onClick={handleSave} disabled={amount < baseAmount}>
            Save
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

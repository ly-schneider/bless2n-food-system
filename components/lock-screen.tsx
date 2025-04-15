"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Lock, Delete } from "lucide-react";

interface LockScreenProps {
  isOpen: boolean;
  onUnlock: () => void;
}

export function LockScreen({ isOpen, onUnlock }: LockScreenProps) {
  const [pin, setPin] = useState("");
  const [error, setError] = useState(false);
  
  // Reset error state when pin changes
  useEffect(() => {
    if (error) setError(false);
    if (pin.length === 4) {
      handleSubmit();
    }
  }, [pin]);

  const handleSubmit = () => {
    // Compare with environment variable
    const correctPin = process.env.NEXT_PUBLIC_UNLOCK_PIN;
    
    if (pin === correctPin) {
      setPin("");
      onUnlock();
    } else {
      setError(true);
      setPin("");
    }
  };

  const addDigit = (digit: string) => {
    // Limit to 4 digits
    if (pin.length < 4) {
      setPin(prev => prev + digit);
    }
  };

  const removeDigit = () => {
    setPin(prev => prev.slice(0, -1));
  };

  return (
    <Dialog open={isOpen} onOpenChange={() => {}}>
      <DialogContent className="rounded-none flex flex-col items-center justify-center bg-background max-w-none w-full min-h-screen">
        <DialogHeader className="text-center">
          <DialogTitle className="text-2xl flex items-center justify-center gap-2">
            <Lock className="h-6 w-6" />
            Display gesperrt
          </DialogTitle>
        </DialogHeader>

        <div className="w-full max-w-xs mx-auto mt-4">
          {/* PIN Display */}
          <div className="text-center mb-6">
            <div className="flex justify-center items-center gap-2">
              {[...Array(4)].map((_, i) => (
                <div 
                  key={i}
                  className={`w-10 h-14 border-2 rounded-md flex items-center justify-center text-xl font-medium ${
                    error ? "border-destructive" : pin.length > i ? "border-primary" : "border-muted-foreground"
                  }`}
                >
                  {pin.length > i ? '•' : ''}
                </div>
              ))}
            </div>
            {error && <p className="text-destructive text-center mt-2">Incorrect PIN</p>}
          </div>
          
          {/* Numpad */}
          <div className="grid grid-cols-3 gap-3">
            {[1, 2, 3, 4, 5, 6, 7, 8, 9].map(num => (
              <Button
                key={num}
                type="button"
                variant="outline"
                className="h-16 text-xl font-medium"
                onClick={() => addDigit(num.toString())}
              >
                {num}
              </Button>
            ))}
            <Button
              type="button"
              variant="outline"
              className="h-16 invisible"
              onClick={removeDigit}
            >
              <Delete className="h-5 w-5" />
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-16 text-xl font-medium"
              onClick={() => addDigit("0")}
            >
              0
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-16"
              onClick={removeDigit}
            >
              <Delete className="h-5 w-5" />
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
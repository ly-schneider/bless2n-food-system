"use client";

import { createContext, useContext, useState, useEffect } from "react";
import { LockScreen } from "./lock-screen";
import { Button } from "@/components/ui/button";
import { Lock } from "lucide-react";

interface LockContextType {
  isLocked: boolean;
  lockApp: () => void;
  unlockApp: () => void;
}

const LockContext = createContext<LockContextType | undefined>(undefined);

export function useLock() {
  const context = useContext(LockContext);
  if (context === undefined) {
    throw new Error("useLock must be used within a LockProvider");
  }
  return context;
}

export function LockProvider({ children }: { children: React.ReactNode }) {
  const [isLocked, setIsLocked] = useState(false);

  // Check if the app should be locked on load
  useEffect(() => {
    const storedLockState = localStorage.getItem("appLocked");
    if (storedLockState === "true") {
      setIsLocked(true);
    }
  }, []);

  // Update localStorage when lock state changes
  useEffect(() => {
    localStorage.setItem("appLocked", isLocked.toString());
  }, [isLocked]);

  const lockApp = () => {
    setIsLocked(true);
  };

  const unlockApp = () => {
    setIsLocked(false);
  };

  return (
    <LockContext.Provider value={{ isLocked, lockApp, unlockApp }}>
      {children}
      <LockScreen isOpen={isLocked} onUnlock={unlockApp} />
    </LockContext.Provider>
  );
}

export function LockButton() {
  const { lockApp } = useLock();

  return (
    <Button 
      variant="outline" 
      onClick={lockApp} 
      aria-label="Lock application"
    >
      <Lock className="h-5 w-5" />
      Sperren
    </Button>
  );
}
"use client";

import { createContext, useContext, useState, useEffect, useCallback, useRef } from "react";
import { LockScreen } from "./lock-screen";
import { Button } from "@/components/ui/button";
import { Lock } from "lucide-react";
import { ActivityTracker } from "./activity-tracker";

interface LockContextType {
  isLocked: boolean;
  lockApp: () => void;
  unlockApp: () => void;
  resetInactivityTimer: () => void;
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
  // Use a ref for the timer so we avoid state updates just for the timer ID.
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const INACTIVITY_TIMEOUT = process.env.NEXT_PUBLIC_INACTIVITY_TIMEOUT
    ? parseInt(process.env.NEXT_PUBLIC_INACTIVITY_TIMEOUT)
    : 300000; // 5 minutes in milliseconds

  // Lock the app and clear the timer
  const lockApp = useCallback(() => {
    setIsLocked(true);
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // Start or reset the inactivity timer
  const startInactivityTimer = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      lockApp();
    }, INACTIVITY_TIMEOUT);
  }, [INACTIVITY_TIMEOUT, lockApp]);

  const resetInactivityTimer = useCallback(() => {
    if (!isLocked) {
      startInactivityTimer();
    }
  }, [isLocked, startInactivityTimer]);

  // Check lock state on load and start timer if not locked
  useEffect(() => {
    const storedLockState = localStorage.getItem("appLocked");
    if (storedLockState === "true") {
      setIsLocked(true);
    } else {
      startInactivityTimer();
    }
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [startInactivityTimer]);

  // Sync lock state with localStorage
  useEffect(() => {
    localStorage.setItem("appLocked", isLocked.toString());
  }, [isLocked]);

  const unlockApp = useCallback(() => {
    setIsLocked(false);
    startInactivityTimer();
  }, [startInactivityTimer]);

  return (
    <LockContext.Provider value={{ isLocked, lockApp, unlockApp, resetInactivityTimer }}>
      {children}
      <ActivityTracker />
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
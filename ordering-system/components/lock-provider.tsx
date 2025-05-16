"use client";

import { createContext, useContext, useState, useEffect, useCallback, useRef } from "react";
import { LockScreen } from "./lock-screen";
import { SleepScreen } from "./sleep-screen";
import { Button } from "@/components/ui/button";
import { Lock, Moon } from "lucide-react";
import { ActivityTracker } from "./activity-tracker";

interface LockContextType {
  isLocked: boolean;
  isAsleep: boolean;
  lockApp: () => void;
  unlockApp: () => void;
  sleepApp: () => void;
  wakeApp: () => void;
  resetInactivityTimer: () => void;
}

const LockContext = createContext<LockContextType | undefined>(undefined);

export function useLock() {
  const context = useContext(LockContext);
  if (!context) {
    throw new Error("useLock must be used within a LockProvider");
  }
  return context;
}

export function LockProvider({ children }: { children: React.ReactNode }) {
  const [isLocked, setIsLocked] = useState(false);
  const [isAsleep, setIsAsleep] = useState(false);
  // Use a ref for the timer so we avoid state updates just for the timer ID.
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const INACTIVITY_TIMEOUT = process.env.NEXT_PUBLIC_INACTIVITY_TIMEOUT
    ? parseInt(process.env.NEXT_PUBLIC_INACTIVITY_TIMEOUT)
    : 300000; // 5 minutes in milliseconds

  // Lock the app and clear the timer
  const lockApp = useCallback(() => {
    setIsLocked(true);
    setIsAsleep(false); // Exit sleep mode when locking
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // Put the app to sleep
  const sleepApp = useCallback(() => {
    setIsAsleep(true);
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // Wake the app from sleep
  const wakeApp = useCallback(() => {
    setIsAsleep(false);
    if (!isLocked) {
      startInactivityTimer();
    }
  }, [isLocked]);

  // Start or reset the inactivity timer
  const startInactivityTimer = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      lockApp();
    }, INACTIVITY_TIMEOUT);
  }, [INACTIVITY_TIMEOUT, lockApp]);

  const resetInactivityTimer = useCallback(() => {
    if (!isLocked && !isAsleep) {
      startInactivityTimer();
    }
  }, [isLocked, isAsleep, startInactivityTimer]);

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
    <LockContext.Provider value={{ 
      isLocked, 
      isAsleep,
      lockApp, 
      unlockApp, 
      sleepApp,
      wakeApp,
      resetInactivityTimer 
    }}>
      {children}
      <ActivityTracker />
      <LockScreen isOpen={isLocked} onUnlock={unlockApp} />
      <SleepScreen isAsleep={isAsleep} onWakeUp={wakeApp} />
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

export function SleepButton() {
  const { sleepApp } = useLock();

  return (
    <Button 
      variant="outline" 
      onClick={sleepApp} 
      size={"lg"}
      aria-label="Put application to sleep"
    >
      <Moon className="h-5 w-5" />
      Ruhemodus
    </Button>
  );
}
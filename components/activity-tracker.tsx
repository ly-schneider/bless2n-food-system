"use client";

import { useEffect } from "react";
import { useLock } from "./lock-provider";

export function ActivityTracker() {
  const { resetInactivityTimer } = useLock();

  useEffect(() => {
    // iPad-specific events
    const iPadActivityEvents = [
      "touchstart",
      "touchmove",
      "touchend",
      "gesturestart",
      "gesturechange",
      "gestureend",
    ];
    
    // Computer-specific events
    const computerActivityEvents = [
      "mousedown",
      "mousemove",
      "keypress",
      "click",
      "scroll",
    ];
    
    // Combined events list
    const allActivityEvents = [...iPadActivityEvents, ...computerActivityEvents];

    const handleUserActivity = () => {
      resetInactivityTimer();
    };

    // Add event listeners for all activity events
    allActivityEvents.forEach((eventName) => {
      window.addEventListener(eventName, handleUserActivity, { passive: true });
    });

    // Cleanup function to remove event listeners
    return () => {
      allActivityEvents.forEach((eventName) => {
        window.removeEventListener(eventName, handleUserActivity);
      });
    };
  }, [resetInactivityTimer]);

  // This component doesn't render anything
  return null;
}
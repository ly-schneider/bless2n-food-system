'use client'

import { useState, useEffect } from 'react'

export function ClientTime() {
  // State to hold the current time as a string
  const [time, setTime] = useState('')

  useEffect(() => {
    // Function to update the time state with the current local time
    const updateTime = () => {
      const now = new Date()
      setTime(now.toLocaleTimeString())
    }
    updateTime() // Initialize time immediately
    const interval = setInterval(updateTime, 1000) // Update time every second
    return () => clearInterval(interval) // Cleanup interval on component unmount
  }, [])

  return <span>{time}</span>
}
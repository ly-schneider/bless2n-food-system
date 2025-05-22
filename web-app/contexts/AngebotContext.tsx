'use client'

import { createContext, useState, useContext, ReactNode } from 'react'

interface AngebotContextType {
  angebot: "lite" | "pro"
  setAngebot: (angebot: "lite" | "pro") => void
}

const AngebotContext = createContext<AngebotContextType | undefined>(undefined)

export function AngebotProvider({ children }: { children: ReactNode }) {
  const [angebot, setAngebot] = useState<AngebotContextType["angebot"]>("lite")

  return (
    <AngebotContext.Provider value={{ angebot, setAngebot }}>
      {children}
    </AngebotContext.Provider>
  )
}

export function useAngebot() {
  const context = useContext(AngebotContext)
  if (!context) throw new Error('useAngebot must be used within an AngebotProvider')
  return context
}
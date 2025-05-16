import { createClientComponentClient } from '@supabase/auth-helpers-nextjs'
import { useEffect, useState, useCallback } from 'react'
import type { User as SupabaseUser } from '@supabase/supabase-js'

export function useSupabase() {
  const [supabase] = useState(() => createClientComponentClient())
  const [user, setUser] = useState<SupabaseUser | null>(null)

  useEffect(() => {
    const { data: { subscription } } = supabase.auth.onAuthStateChange((event, session) => {
      console.log("Auth state changed:", event, session)
      if (session) {
        setUser(session.user)
      } else {
        setUser(null)
      }
    })

    // Check the initial session
    supabase.auth.getSession().then(({ data: { session } }) => {
      console.log("Initial session:", session)
      setUser(session?.user ?? null)
    })

    return () => {
      subscription.unsubscribe()
    }
  }, [supabase])

  const signOut = useCallback(async () => {
    try {
      const { error } = await supabase.auth.signOut()
      if (error) throw error
    } catch (error) {
      console.error('Error signing out:', error)
      throw error
    }
  }, [supabase])

  return { supabase, user, setUser, signOut }
}

import { createContext, FC, PropsWithChildren, useContext, useMemo } from "react"
import { DefaultTheme, Theme as NavTheme } from "@react-navigation/native"
import * as SystemUI from "expo-system-ui"
import { StyleProp, StyleSheet } from "react-native"

import { theme, ThemedStyle, ThemedStyleArray } from "@/theme"

type ThemeContextType = { theme: typeof theme }
const ThemeContext = createContext<ThemeContextType>({ theme })

export const ThemeProvider: FC<PropsWithChildren> = ({ children }) => {
  SystemUI.setBackgroundColorAsync(theme.colors.background)
  const value = useMemo(() => ({ theme }), [])
  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export const useNavTheme = (): NavTheme =>
  useMemo(
    () => ({
      ...DefaultTheme,
      colors: { ...DefaultTheme.colors, ...theme.colors },
    }),
    [],
  )

export const useAppTheme = () => {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error("useAppTheme must be used inside <ThemeProvider>")

  const themed = <T extends object>(
    input: ThemedStyle<T> | StyleProp<T> | ThemedStyleArray<T>,
  ): T => {
    const resolved = [input]
      .flat(Infinity)
      .map((p) => (typeof p === "function" ? (p as ThemedStyle<T>)(theme) : p))
    return StyleSheet.flatten(resolved) as T
  }

  return { theme: ctx.theme, navTheme: useNavTheme(), themed }
}

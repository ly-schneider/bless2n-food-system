import type { StyleProp } from "react-native"
import { timing } from "./timing"
import { typography } from "./typography"
import { colors } from "./colors"
import { spacing } from "./spacing"

export type Colors = typeof colors
export type Spacing = typeof spacing
export type Timing = typeof timing
export type Typography = typeof typography
export interface Theme {
  colors: Colors
  spacing: Spacing
  typography: Typography
  timing: Timing
}

export const theme: Theme = {
  colors,
  spacing,
  typography,
  timing,
}

export type ThemedStyle<T> = (theme: Theme) => T
export type ThemedStyleArray<T> = (
  | ThemedStyle<T>
  | StyleProp<T>
  | (StyleProp<T> | ThemedStyle<T>)[]
)[]

export * from "./styles"
export * from "./typography"
export * from "./timing"

import { FC } from "react"
import { View, ViewStyle, TextStyle } from "react-native"
import { Screen } from "@/components/Screen"
import { Text } from "@/components/Text"
import { useAppTheme } from "@/utils/useAppTheme"
import { type ThemedStyle } from "@/theme"
import type { AppStackScreenProps } from "@/navigators/AppNavigator"

type HomeScreenProps = AppStackScreenProps<"Home">

export const HomeScreen: FC<HomeScreenProps> = () => {
  const { themed } = useAppTheme()

  return (
    <Screen preset="scroll" style={themed($container)}>
      <View style={themed($header)}>
        <Text tx="homeScreen:title" preset="heading" style={themed($title)} />
      </View>
    </Screen>
  )
}

const $container: ThemedStyle<ViewStyle> = ({ colors }) => ({
  flex: 1,
  backgroundColor: colors.background,
})

const $header: ThemedStyle<ViewStyle> = ({ spacing }) => ({
  paddingHorizontal: spacing.lg,
  paddingTop: spacing.xl,
  paddingBottom: spacing.lg,
  alignItems: "center",
})

const $title: ThemedStyle<TextStyle> = ({ colors, typography }) => ({
  color: colors.text,
  fontFamily: typography.heading.fontFamily,
  fontSize: 32,
  textAlign: "center",
})

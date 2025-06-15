import { FC, ReactNode } from "react"
import { Pressable, ViewStyle, TextStyle } from "react-native"
import Animated, { useSharedValue, useAnimatedStyle, withSpring } from "react-native-reanimated"
import { Text, type TextProps } from "./Text"
import { useAppTheme } from "@/utils/useAppTheme"
import { type ThemedStyle } from "@/theme"

const AnimatedPressable = Animated.createAnimatedComponent(Pressable)

export interface PrimaryButtonProps {
  /**
   * The text to display on the button
   */
  text?: string
  /**
   * Text translation key
   */
  tx?: TextProps["tx"]
  /**
   * Optional options to pass to the translation
   */
  txOptions?: TextProps["txOptions"]
  /**
   * Children to render instead of text
   */
  children?: ReactNode
  /**
   * Whether the button is disabled
   */
  disabled?: boolean
  /**
   * Callback when the button is pressed
   */
  onPress?: () => void
  /**
   * Additional styling for the container
   */
  className?: string
  /**
   * Additional styling for the container
   */
  style?: ViewStyle
  /**
   * Additional styling for the text
   */
  textStyle?: TextStyle
}

/**
 * PrimaryButton component with coral pink background and one-thumb size.
 * Features press animations and disabled state.
 */
export const PrimaryButton: FC<PrimaryButtonProps> = ({
  text,
  tx,
  txOptions,
  children,
  disabled = false,
  onPress,
  style,
  textStyle,
}) => {
  const { themed } = useAppTheme()
  const scale = useSharedValue(1)

  const animatedStyle = useAnimatedStyle(() => ({
    transform: [{ scale: scale.value }],
  }))

  const handlePressIn = () => {
    if (!disabled) {
      scale.value = withSpring(0.96, { damping: 15, stiffness: 300 })
    }
  }

  const handlePressOut = () => {
    if (!disabled) {
      scale.value = withSpring(1, { damping: 15, stiffness: 300 })
    }
  }

  const handlePress = () => {
    if (!disabled && onPress) {
      onPress()
    }
  }

  const content = children || (
    <Text
      tx={tx}
      text={text}
      txOptions={txOptions}
      style={[themed($buttonText), disabled && themed($disabledText), textStyle]}
    />
  )

  return (
    <AnimatedPressable
      style={[themed($container), disabled && themed($disabledContainer), animatedStyle, style]}
      onPressIn={handlePressIn}
      onPressOut={handlePressOut}
      onPress={handlePress}
      disabled={disabled}
    >
      {content}
    </AnimatedPressable>
  )
}

const $container: ThemedStyle<ViewStyle> = ({ colors, spacing }) => ({
  backgroundColor: colors.text,
  borderRadius: 28,
  paddingVertical: spacing.md,
  paddingHorizontal: spacing.lg,
  justifyContent: "center",
  alignItems: "center",
  minHeight: 56,
  shadowColor: colors.background,
  shadowOffset: { width: 0, height: 4 },
  shadowOpacity: 0.2,
  shadowRadius: 8,
  elevation: 4,
})

const $disabledContainer: ThemedStyle<ViewStyle> = ({ colors }) => ({
  backgroundColor: colors.textDim,
  opacity: 0.6,
})

const $buttonText: ThemedStyle<TextStyle> = ({ colors, typography }) => ({
  color: colors.background,
  fontFamily: typography.primary.regular,
  fontSize: 18,
  lineHeight: 24,
  fontWeight: "bold",
  textAlign: "center",
})

const $disabledText: ThemedStyle<TextStyle> = ({ colors }) => ({
  color: colors.textDim,
})

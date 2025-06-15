import { FC } from "react"
import { View, ViewStyle, TextStyle, TouchableOpacity } from "react-native"
import { Text, TextProps } from "./Text"
import { useAppTheme } from "@/utils/useAppTheme"
import { type ThemedStyle } from "@/theme"

export interface NumberSliderProps {
  /**
   * Label to display above the slider
   */
  label?: string
  /**
   * Label text which is looked up via i18n.
   */
  tx?: TextProps["tx"]
  /**
   * Optional label options to pass to i18n. Useful for interpolation
   * as well as explicitly setting locale or translation fallbacks.
   */
  txOptions?: TextProps["txOptions"]
  /**
   * Current value of the slider
   */
  value: number
  /**
   * Minimum value
   */
  minimumValue: number
  /**
   * Maximum value
   */
  maximumValue: number
  /**
   * Step size for slider values
   */
  step?: number
  /**
   * Callback when value changes
   */
  onValueChange: (value: number) => void
  /**
   * Optional style override for the container
   */
  style?: ViewStyle
}

export const NumberSlider: FC<NumberSliderProps> = ({
  label,
  tx,
  txOptions,
  value,
  minimumValue,
  maximumValue,
  step = 1,
  onValueChange,
  style,
}) => {
  const { themed } = useAppTheme()

  const handleDecrease = () => {
    const newValue = Math.max(minimumValue, value - step)
    onValueChange(newValue)
  }

  const handleIncrease = () => {
    const newValue = Math.min(maximumValue, value + step)
    onValueChange(newValue)
  }

  return (
    <View style={[themed($container), style]}>
      <View style={themed($labelContainer)}>
        <Text
          text={label}
          tx={tx}
          txOptions={txOptions}
          preset="formLabel"
          style={themed($label)}
        />
        <Text text={value.toString()} preset="bold" style={themed($value)} />
      </View>

      <View style={themed($sliderContainer)}>
        <TouchableOpacity
          style={themed($button)}
          onPress={handleDecrease}
          disabled={value <= minimumValue}
        >
          <Text text="-" style={themed($buttonText)} />
        </TouchableOpacity>

        <View style={themed($trackContainer)}>
          <View style={themed($track)} />
          <View
            style={[
              themed($activeTrack),
              { width: `${((value - minimumValue) / (maximumValue - minimumValue)) * 100}%` },
            ]}
          />
          <View
            style={[
              themed($thumb),
              { left: `${((value - minimumValue) / (maximumValue - minimumValue)) * 100}%` },
            ]}
          />
        </View>

        <TouchableOpacity
          style={themed($button)}
          onPress={handleIncrease}
          disabled={value >= maximumValue}
        >
          <Text text="+" style={themed($buttonText)} />
        </TouchableOpacity>
      </View>
    </View>
  )
}

const $container: ThemedStyle<ViewStyle> = ({ spacing }) => ({
  marginVertical: spacing.sm,
})

const $labelContainer: ThemedStyle<ViewStyle> = ({ spacing }) => ({
  flexDirection: "row",
  justifyContent: "space-between",
  alignItems: "center",
  marginBottom: spacing.xs,
})

const $label: ThemedStyle<TextStyle> = ({ colors, typography }) => ({
  color: colors.text,
  fontFamily: typography.primary.regular,
  fontSize: 16,
})

const $value: ThemedStyle<TextStyle> = ({ colors, typography }) => ({
  color: colors.text,
  fontFamily: typography.primary.regular,
  fontSize: 16,
})

const $sliderContainer: ThemedStyle<ViewStyle> = ({ spacing }) => ({
  flexDirection: "row",
  alignItems: "center",
  marginTop: spacing.sm,
})

const $button: ThemedStyle<ViewStyle> = ({ colors, spacing }) => ({
  width: 40,
  height: 40,
  backgroundColor: colors.text,
  borderRadius: 20,
  justifyContent: "center",
  alignItems: "center",
  marginHorizontal: spacing.xs,
})

const $buttonText: ThemedStyle<TextStyle> = ({ colors, typography }) => ({
  color: colors.background,
  fontFamily: typography.primary.regular,
  fontSize: 20,
})

const $trackContainer: ThemedStyle<ViewStyle> = () => ({
  flex: 1,
  height: 40,
  justifyContent: "center",
  position: "relative",
})

const $track: ThemedStyle<ViewStyle> = ({ colors }) => ({
  height: 4,
  backgroundColor: colors.background,
  borderRadius: 2,
})

const $activeTrack: ThemedStyle<ViewStyle> = ({ colors }) => ({
  position: "absolute",
  height: 4,
  backgroundColor: colors.text,
  borderRadius: 2,
})

const $thumb: ThemedStyle<ViewStyle> = ({ colors }) => ({
  position: "absolute",
  width: 20,
  height: 20,
  backgroundColor: colors.text,
  borderRadius: 10,
  marginLeft: -10,
  marginTop: -8,
})

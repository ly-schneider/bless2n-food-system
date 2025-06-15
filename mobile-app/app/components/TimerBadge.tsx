import { FC, useState, useEffect } from "react"
import { View, ViewStyle, TextStyle } from "react-native"
import { Text } from "./Text"
import { useAppTheme } from "@/utils/useAppTheme"
import { type ThemedStyle } from "@/theme"

export interface TimerBadgeProps {
  /**
   * Initial time in seconds
   */
  initialSeconds: number
  /**
   * Whether the timer is running
   */
  isRunning?: boolean
  /**
   * Callback when timer reaches zero
   */
  onComplete?: () => void
  /**
   * Callback for each second that passes
   */
  onTick?: (remainingSeconds: number) => void
  /**
   * Whether to show warning style when time is low
   */
  showWarning?: boolean
  /**
   * Threshold in seconds for warning style (default: 30)
   */
  warningThreshold?: number
  /**
   * Optional style override for the container
   */
  style?: ViewStyle
}

export const TimerBadge: FC<TimerBadgeProps> = ({
  initialSeconds,
  isRunning = true,
  onComplete,
  onTick,
  showWarning = true,
  warningThreshold = 30,
  style,
}) => {
  const { themed } = useAppTheme()
  const [remainingSeconds, setRemainingSeconds] = useState(initialSeconds)

  useEffect(() => {
    setRemainingSeconds(initialSeconds)
  }, [initialSeconds])

  useEffect(() => {
    if (!isRunning || remainingSeconds <= 0) return

    const timer = setInterval(() => {
      setRemainingSeconds((prev) => {
        const newTime = prev - 1

        if (onTick) {
          onTick(newTime)
        }

        if (newTime <= 0) {
          if (onComplete) {
            onComplete()
          }
          return 0
        }

        return newTime
      })
    }, 1000)

    return () => clearInterval(timer)
  }, [isRunning, remainingSeconds, onComplete, onTick])

  const formatTime = (seconds: number): string => {
    const minutes = Math.floor(seconds / 60)
    const remainderSeconds = seconds % 60
    return `${minutes}:${remainderSeconds.toString().padStart(2, "0")}`
  }

  const isWarning = showWarning && remainingSeconds <= warningThreshold && remainingSeconds > 0
  const isExpired = remainingSeconds <= 0

  return (
    <View
      style={[
        themed($container),
        isWarning ? themed($containerWarning) : undefined,
        isExpired ? themed($containerExpired) : undefined,
        style,
      ]}
    >
      <Text text="⏱️" style={themed($icon)} />
      <Text
        text={formatTime(remainingSeconds)}
        style={[
          themed($timerText),
          isWarning ? themed($timerTextWarning) : undefined,
          isExpired ? themed($timerTextExpired) : undefined,
        ]}
      />
    </View>
  )
}

const $container: ThemedStyle<ViewStyle> = ({ colors, spacing }) => ({
  flexDirection: "row",
  alignItems: "center",
  backgroundColor: colors.background,
  paddingHorizontal: spacing.md,
  paddingVertical: spacing.sm,
  borderRadius: 20,
  borderWidth: 2,
  borderColor: colors.border,
})

const $containerWarning: ThemedStyle<ViewStyle> = ({ colors }) => ({
  backgroundColor: colors.text,
  borderColor: colors.secondary,
})

const $containerExpired: ThemedStyle<ViewStyle> = ({ colors }) => ({
  backgroundColor: `${colors.error}40`,
  borderColor: colors.error,
})

const $icon: TextStyle = {
  fontSize: 16,
  marginRight: 6,
}

const $timerText: ThemedStyle<TextStyle> = ({ colors, typography }) => ({
  color: colors.text,
  fontFamily: typography.primary.regular,
  fontSize: 18,
})

const $timerTextWarning: ThemedStyle<TextStyle> = ({ colors }) => ({
  color: colors.secondary,
})

const $timerTextExpired: ThemedStyle<TextStyle> = ({ colors }) => ({
  color: colors.error,
})

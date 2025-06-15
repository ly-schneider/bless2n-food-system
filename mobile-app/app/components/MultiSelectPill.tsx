import { FC } from "react"
import { View, ViewStyle, TextStyle, TouchableOpacity, FlatList } from "react-native"
import { Text, TextProps } from "./Text"
import { useAppTheme } from "@/utils/useAppTheme"
import { type ThemedStyle } from "@/theme"

export interface PillOption {
  id: string
  label: string
  icon?: string
}

export interface MultiSelectPillProps {
  /**
   * Label to display above the pills
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
   * Array of options to display as pills
   */
  options: PillOption[]
  /**
   * Array of selected option IDs
   */
  selectedIds: string[]
  /**
   * Callback when selection changes
   */
  onSelectionChange: (selectedIds: string[]) => void
  /**
   * Allow multiple selections (default: true)
   */
  multiple?: boolean
  /**
   * Optional style override for the container
   */
  style?: ViewStyle
}

export const MultiSelectPill: FC<MultiSelectPillProps> = ({
  label,
  tx,
  txOptions,
  options,
  selectedIds,
  onSelectionChange,
  multiple = true,
  style,
}) => {
  const { themed } = useAppTheme()

  const handlePillPress = (optionId: string) => {
    if (multiple) {
      const isSelected = selectedIds.includes(optionId)
      if (isSelected) {
        onSelectionChange(selectedIds.filter((id) => id !== optionId))
      } else {
        onSelectionChange([...selectedIds, optionId])
      }
    } else {
      onSelectionChange([optionId])
    }
  }

  const renderPill = ({ item }: { item: PillOption }) => {
    const isSelected = selectedIds.includes(item.id)

    return (
      <TouchableOpacity
        style={[themed($pill), isSelected ? themed($pillSelected) : themed($pillUnselected)]}
        onPress={() => handlePillPress(item.id)}
        activeOpacity={0.7}
      >
        {item.icon && <Text text={item.icon} style={themed($pillIcon)} />}
        <Text
          text={item.label}
          style={[
            themed($pillText),
            isSelected ? themed($pillTextSelected) : themed($pillTextUnselected),
          ]}
        />
      </TouchableOpacity>
    )
  }

  return (
    <View style={[themed($container), style]}>
      {(label || tx) && (
        <Text
          text={label}
          tx={tx}
          txOptions={txOptions}
          preset="formLabel"
          style={themed($label)}
        />
      )}
      <FlatList
        data={options}
        renderItem={renderPill}
        keyExtractor={(item) => item.id}
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={themed($pillContainer)}
      />
    </View>
  )
}

const $container: ThemedStyle<ViewStyle> = ({ spacing }) => ({
  marginVertical: spacing.sm,
})

const $label: ThemedStyle<TextStyle> = ({ colors, typography, spacing }) => ({
  color: colors.text,
  fontFamily: typography.primary.regular,
  fontSize: 16,
  marginBottom: spacing.xs,
})

const $pillContainer: ThemedStyle<ViewStyle> = ({ spacing }) => ({
  paddingVertical: spacing.xs,
})

const $pill: ThemedStyle<ViewStyle> = ({ spacing }) => ({
  flexDirection: "row",
  alignItems: "center",
  paddingHorizontal: spacing.md,
  paddingVertical: spacing.sm,
  marginRight: spacing.sm,
  borderRadius: 20,
  borderWidth: 2,
})

const $pillSelected: ThemedStyle<ViewStyle> = ({ colors }) => ({
  backgroundColor: colors.text,
  borderColor: colors.text,
})

const $pillUnselected: ThemedStyle<ViewStyle> = ({ colors }) => ({
  backgroundColor: "transparent",
  borderColor: colors.border,
})

const $pillIcon: TextStyle = {
  marginRight: 4,
  fontSize: 16,
}

const $pillText: ThemedStyle<TextStyle> = ({ typography }) => ({
  fontFamily: typography.primary.regular,
  fontSize: 14,
})

const $pillTextSelected: ThemedStyle<TextStyle> = ({ colors }) => ({
  color: colors.background,
})

const $pillTextUnselected: ThemedStyle<TextStyle> = ({ colors }) => ({
  color: colors.text,
})

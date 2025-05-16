import React, { FC, useRef, useEffect } from 'react';
import {
  Animated, Pressable,
  StyleProp,
  ViewStyle,
  PressableProps
} from 'react-native';
import * as Haptics from 'expo-haptics';

interface AnimatedHapticButtonProps extends Omit<PressableProps, 'style'> {
  disabled?: boolean;
  onPress: () => void;
  children: React.ReactNode;
  style?: StyleProp<ViewStyle>;
  pressableStyle?: StyleProp<ViewStyle>;
  useDefaultStyles?: boolean;
  useAnimation?: boolean;
  useHaptics?: boolean;
  animationDuration?: number;
}

export const AnimatedHapticButton: FC<AnimatedHapticButtonProps> = ({
  disabled = false,
  onPress,
  children,
  style,
  pressableStyle,
  useDefaultStyles = true,
  useAnimation = true,
  useHaptics = true,
  animationDuration = 200,
  ...pressableProps
}) => {
  const opacity = useRef(new Animated.Value(disabled ? 0.5 : 1)).current;

  useEffect(() => {
    if (useAnimation) {
      Animated.timing(opacity, {
        toValue: disabled ? 0.5 : 1,
        duration: animationDuration,
        useNativeDriver: true,
      }).start();
    }
  }, [disabled, useAnimation, animationDuration]);

  const handlePressIn = () => {
    if (!disabled && useHaptics && process.env.EXPO_OS === 'ios') {
      Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Soft);
    }
  };

  const containerStyle = useAnimation 
    ? [style, { opacity }]
    : [style, disabled ? { opacity: 0.5 } : undefined];

  return (
    <Animated.View style={containerStyle}>
      <Pressable
        style={[useDefaultStyles && { flex: 1, width: '100%', height: '100%', justifyContent: "center", alignItems: "center" }, pressableStyle]}
        disabled={disabled}
        onPressIn={handlePressIn}
        onPress={onPress}
        {...pressableProps}
      >
        {children}
      </Pressable>
    </Animated.View>
  );
};

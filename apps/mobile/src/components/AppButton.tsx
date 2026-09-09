import React from 'react';
import {ActivityIndicator, Pressable, StyleSheet, Text, ViewStyle} from 'react-native';
import {colors, control, radius, spacing} from '../theme/tokens';

type Variant = 'primary' | 'secondary' | 'danger' | 'text';

export function AppButton({
  label,
  onPress,
  testID,
  accessibilityLabel = label,
  variant = 'primary',
  disabled = false,
  loading = false,
  style,
}: {
  label: string;
  onPress: () => void;
  testID?: string;
  accessibilityLabel?: string;
  variant?: Variant;
  disabled?: boolean;
  loading?: boolean;
  style?: ViewStyle;
}) {
  const inactive = disabled || loading;
  const stableTestID = testID ?? `button-${label.toLowerCase().replace(/[^a-zа-яё0-9]+/giu, '-').replace(/^-|-$/g, '')}`;
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={accessibilityLabel}
      accessibilityState={{disabled: inactive, busy: loading}}
      testID={stableTestID}
      disabled={inactive}
      hitSlop={4}
      onPress={onPress}
      style={({pressed}) => [
        styles.base,
        styles[variant],
        inactive && styles.disabled,
        pressed && !inactive && styles.pressed,
        style,
      ]}>
      {loading ? (
        <ActivityIndicator color={variant === 'primary' ? colors.inverse : colors.text} />
      ) : (
        <Text style={[styles.label, variant === 'primary' && styles.primaryLabel, variant === 'danger' && styles.dangerLabel]}>{label}</Text>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: {
    minHeight: control.minTouch,
    borderRadius: radius.md,
    paddingHorizontal: spacing.md,
    alignItems: 'center',
    justifyContent: 'center',
  },
  primary: {backgroundColor: colors.primary},
  secondary: {borderWidth: 1, borderColor: colors.border, backgroundColor: colors.surface},
  danger: {borderWidth: 1, borderColor: colors.danger, backgroundColor: colors.surface},
  text: {backgroundColor: 'transparent'},
  label: {fontSize: 15, fontWeight: '800', color: colors.text},
  primaryLabel: {color: colors.inverse},
  dangerLabel: {color: colors.danger},
  disabled: {opacity: 0.45},
  pressed: {opacity: 0.72},
});

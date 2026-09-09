import React from 'react';
import {Pressable, StyleSheet, Text, View} from 'react-native';
import {colors, control, spacing} from '../theme/tokens';

export type MainTab = 'home' | 'training' | 'nutrition' | 'progress' | 'ai';

const tabs: Array<{id: MainTab; label: string; glyph: string}> = [
  {id: 'home', label: 'Сегодня', glyph: '●'},
  {id: 'training', label: 'Тренировка', glyph: '◆'},
  {id: 'nutrition', label: 'Питание', glyph: '◐'},
  {id: 'progress', label: 'Прогресс', glyph: '↗'},
  {id: 'ai', label: 'AI', glyph: '✦'},
];

export function BottomNavigation({current, onSelect}: {current: MainTab; onSelect: (tab: MainTab) => void}) {
  return (
    <View style={styles.container} accessibilityRole="tablist">
      {tabs.map(tab => {
        const selected = current === tab.id;
        return (
          <Pressable
            key={tab.id}
            accessibilityRole="tab"
            accessibilityLabel={tab.label}
            accessibilityState={{selected}}
            testID={`tab-${tab.id}`}
            hitSlop={2}
            onPress={() => onSelect(tab.id)}
            style={({pressed}) => [styles.item, selected && styles.itemSelected, pressed && styles.pressed]}>
            <Text style={[styles.glyph, selected && styles.selectedText]}>{tab.glyph}</Text>
            <Text numberOfLines={1} style={[styles.label, selected && styles.selectedText]}>{tab.label}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    minHeight: 68,
    borderTopWidth: 1,
    borderTopColor: colors.border,
    backgroundColor: colors.surface,
    flexDirection: 'row',
    paddingHorizontal: spacing.xs,
    paddingTop: spacing.xs,
  },
  item: {flex: 1, minHeight: control.minTouch, alignItems: 'center', justifyContent: 'center', gap: 2, borderRadius: 12},
  itemSelected: {backgroundColor: colors.surfaceMuted},
  pressed: {opacity: 0.65},
  glyph: {fontSize: 17, color: colors.textMuted},
  label: {fontSize: 10, fontWeight: '800', color: colors.textMuted},
  selectedText: {color: colors.text},
});

import React from 'react';
import {Pressable, StyleSheet, Text, View} from 'react-native';
import {colors, control, spacing} from '../theme/tokens';

export type MainTab = 'home' | 'training' | 'nutrition' | 'progress' | 'ai';

const tabs: Array<{id: MainTab; label: string}> = [
  {id: 'home', label: 'Сегодня'},
  {id: 'training', label: 'Тренировка'},
  {id: 'nutrition', label: 'Питание'},
  {id: 'progress', label: 'Прогресс'},
  {id: 'ai', label: 'AI тренер'},
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
            onPress={() => onSelect(tab.id)}
            style={({pressed}) => [styles.item, selected && styles.itemSelected, pressed && styles.pressed]}>
            <Text style={[styles.label, selected && styles.selectedText]}>{tab.label}</Text>
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
    paddingBottom: spacing.xs,
  },
  item: {flex: 1, minHeight: control.minTouch, paddingVertical: spacing.sm, paddingHorizontal: 2, alignItems: 'center', justifyContent: 'center', borderRadius: 12},
  itemSelected: {backgroundColor: colors.surfaceMuted, borderBottomWidth: 3, borderBottomColor: colors.primary},
  pressed: {opacity: 0.65},
  label: {fontSize: 12, lineHeight: 16, fontWeight: '700', textAlign: 'center', color: colors.textMuted},
  selectedText: {color: colors.text},
});

import React, {useState} from 'react';
import {Pressable, StyleSheet, Text, View} from 'react-native';
import {MuscleId, muscleMeta} from '../domain/muscles';
import {colors, control, radius, spacing} from '../theme/tokens';

type Side = 'front' | 'back';
type Region = {label: string; muscles: MuscleId[]};

// Separate regions keep the muscle choices readable and their touch targets independent.
const regions: Record<Side, Region[]> = {
  front: [
    {label: 'Верх тела', muscles: ['shoulders', 'chest', 'biceps', 'forearms']},
    {label: 'Корпус', muscles: ['core']},
    {label: 'Ноги', muscles: ['quads', 'calves']},
  ],
  back: [
    {label: 'Верх тела', muscles: ['shoulders', 'back', 'triceps', 'forearms']},
    {label: 'Таз и ягодицы', muscles: ['glutes']},
    {label: 'Ноги', muscles: ['hamstrings', 'calves']},
  ],
};

export function BodyMap({onSelect}: {onSelect: (muscle: MuscleId) => void}) {
  const [side, setSide] = useState<Side>('front');

  return (
    <View style={styles.wrap} testID="body-map">
      <Text style={styles.eyebrow}>КАРТА ТЕЛА</Text>
      <Text accessibilityRole="header" style={styles.title}>Мышцы по областям</Text>
      <Text style={styles.intro}>Выбери сторону тела и группу мышц для тренировки.</Text>
      <View style={styles.segment} accessibilityRole="tablist">
        {(['front', 'back'] as const).map(value => (
          <Pressable
            key={value}
            accessibilityRole="tab"
            accessibilityLabel={value === 'front' ? 'Вид спереди' : 'Вид сзади'}
            accessibilityState={{selected: side === value}}
            testID={`body-map-side-${value}`}
            onPress={() => setSide(value)}
            style={({pressed}) => [styles.segmentItem, side === value && styles.segmentActive, pressed && styles.pressed]}>
            <Text style={[styles.segmentText, side === value && styles.segmentTextActive]}>{value === 'front' ? 'Спереди' : 'Сзади'}</Text>
          </Pressable>
        ))}
      </View>

      <View
        testID="body-map-illustration"
        accessible={false}
        accessibilityElementsHidden
        importantForAccessibility="no-hide-descendants"
        pointerEvents="none"
        style={styles.silhouettePanel}>
        <View style={styles.figure}>
          <View style={styles.head} />
          <View style={styles.neck} />
          <View style={styles.shoulders} />
          <View style={styles.armLeft} />
          <View style={styles.armRight} />
          <View style={styles.torso} />
          <View style={styles.legLeft} />
          <View style={styles.legRight} />
          <View style={side === 'front' ? styles.frontMark : styles.backMark} />
        </View>
      </View>

      {regions[side].map(region => (
        <View key={region.label} style={styles.region}>
          <Text accessibilityRole="header" style={styles.regionTitle}>{region.label}</Text>
          <View style={styles.muscleGrid}>
            {region.muscles.map(muscle => (
              <Pressable
                key={muscle}
                accessibilityRole="button"
                accessibilityLabel={`Выбрать группу мышц: ${muscleMeta[muscle].title}`}
                testID={`body-map-${side}-${muscle}`}
                onPress={() => onSelect(muscle)}
                style={({pressed}) => [styles.muscleButton, pressed && styles.pressed]}>
                <Text style={styles.muscleText}>{muscleMeta[muscle].title}</Text>
                <Text style={styles.arrow} accessibilityElementsHidden importantForAccessibility="no">→</Text>
              </Pressable>
            ))}
          </View>
        </View>
      ))}
      <Text style={styles.note}>Эта схема помогает выбрать тренировку и не оценивает состояние мышц.</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: spacing.md, gap: spacing.sm, backgroundColor: colors.surface},
  eyebrow: {fontSize: 11, fontWeight: '800', letterSpacing: 1.2, color: colors.textMuted},
  title: {fontSize: 22, fontWeight: '800', color: colors.text},
  intro: {fontSize: 14, lineHeight: 20, color: colors.textMuted},
  segment: {flexDirection: 'row', padding: spacing.xs, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, marginTop: spacing.xs},
  segmentItem: {flex: 1, minHeight: control.minTouch, alignItems: 'center', justifyContent: 'center', borderRadius: radius.sm, paddingHorizontal: spacing.xs},
  segmentActive: {backgroundColor: colors.primary},
  segmentText: {fontWeight: '700', fontSize: 14, color: colors.text},
  segmentTextActive: {color: colors.inverse},
  silhouettePanel: {height: 142, borderRadius: radius.md, backgroundColor: colors.surfaceMuted, alignItems: 'center', justifyContent: 'center', marginTop: spacing.xs},
  figure: {width: 100, height: 134},
  head: {position: 'absolute', top: 2, left: 40, width: 20, height: 23, borderRadius: 12, backgroundColor: colors.textMuted},
  neck: {position: 'absolute', top: 24, left: 45, width: 10, height: 12, backgroundColor: colors.textMuted},
  shoulders: {position: 'absolute', top: 36, left: 19, width: 62, height: 16, borderRadius: 8, backgroundColor: colors.textMuted},
  armLeft: {position: 'absolute', top: 42, left: 18, width: 12, height: 53, borderRadius: 7, backgroundColor: colors.textMuted},
  armRight: {position: 'absolute', top: 42, right: 18, width: 12, height: 53, borderRadius: 7, backgroundColor: colors.textMuted},
  torso: {position: 'absolute', top: 36, left: 30, width: 40, height: 53, borderRadius: 14, backgroundColor: colors.textMuted},
  legLeft: {position: 'absolute', top: 86, left: 33, width: 14, height: 46, borderRadius: 8, backgroundColor: colors.textMuted},
  legRight: {position: 'absolute', top: 86, right: 33, width: 14, height: 46, borderRadius: 8, backgroundColor: colors.textMuted},
  frontMark: {position: 'absolute', top: 54, left: 40, width: 20, height: 3, borderRadius: 2, backgroundColor: colors.inverse},
  backMark: {position: 'absolute', top: 48, left: 49, width: 3, height: 29, borderRadius: 2, backgroundColor: colors.inverse},
  region: {gap: spacing.sm, marginTop: spacing.sm},
  regionTitle: {fontSize: 15, fontWeight: '800', color: colors.text},
  muscleGrid: {flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm},
  muscleButton: {minWidth: '45%', flexGrow: 1, flexBasis: '45%', minHeight: control.minTouch, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, backgroundColor: colors.surfaceMuted, paddingHorizontal: spacing.sm, paddingVertical: spacing.sm, flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: spacing.xs},
  muscleText: {flex: 1, fontSize: 14, fontWeight: '700', color: colors.text},
  arrow: {fontSize: 17, color: colors.text},
  pressed: {opacity: 0.7},
  note: {fontSize: 12, lineHeight: 18, color: colors.textMuted, marginTop: spacing.xs},
});

import React, {useMemo, useState} from 'react';
import {Pressable, StyleSheet, Text, View} from 'react-native';
import {MuscleId, muscleMeta} from '../domain/muscles';

type Side = 'front' | 'back';

type Point = {muscle: MuscleId; top: number; left: number; width?: number};

const points: Record<Side, Point[]> = {
  front: [
    {muscle: 'shoulders', top: 72, left: 18, width: 110},
    {muscle: 'chest', top: 118, left: 57, width: 112},
    {muscle: 'biceps', top: 154, left: 5, width: 92},
    {muscle: 'forearms', top: 216, left: 4, width: 104},
    {muscle: 'core', top: 202, left: 80, width: 72},
    {muscle: 'quads', top: 319, left: 57, width: 118},
    {muscle: 'calves', top: 430, left: 62, width: 105},
  ],
  back: [
    {muscle: 'shoulders', top: 72, left: 18, width: 110},
    {muscle: 'back', top: 125, left: 62, width: 102},
    {muscle: 'triceps', top: 163, left: 4, width: 92},
    {muscle: 'forearms', top: 216, left: 4, width: 104},
    {muscle: 'glutes', top: 275, left: 69, width: 92},
    {muscle: 'hamstrings', top: 344, left: 41, width: 151},
    {muscle: 'calves', top: 430, left: 62, width: 105},
  ],
};

export function BodyMap({onSelect}: {onSelect: (muscle: MuscleId) => void}) {
  const [side, setSide] = useState<Side>('front');
  const sidePoints = useMemo(() => points[side], [side]);

  return (
    <View style={styles.wrap}>
      <View style={styles.header}>
        <View>
          <Text style={styles.eyebrow}>КАРТА ТЕЛА</Text>
          <Text style={styles.title}>Нажми на мышцу</Text>
        </View>
        <View style={styles.segment}>
          <Pressable onPress={() => setSide('front')} style={[styles.segmentItem, side === 'front' && styles.segmentActive]}>
            <Text style={[styles.segmentText, side === 'front' && styles.segmentTextActive]}>Спереди</Text>
          </Pressable>
          <Pressable onPress={() => setSide('back')} style={[styles.segmentItem, side === 'back' && styles.segmentActive]}>
            <Text style={[styles.segmentText, side === 'back' && styles.segmentTextActive]}>Сзади</Text>
          </Pressable>
        </View>
      </View>

      <View style={styles.bodyCanvas}>
        <View style={styles.head} />
        <View style={styles.neck} />
        <View style={styles.torso} />
        <View style={[styles.arm, styles.armLeft]} />
        <View style={[styles.arm, styles.armRight]} />
        <View style={[styles.leg, styles.legLeft]} />
        <View style={[styles.leg, styles.legRight]} />
        {sidePoints.map(point => (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={`Выбрать: ${muscleMeta[point.muscle].title}`}
            key={point.muscle}
            onPress={() => onSelect(point.muscle)}
            style={[styles.hotspot, {top: point.top, left: point.left, width: point.width ?? 100}]}>
            <Text style={styles.hotspotText}>{muscleMeta[point.muscle].short}</Text>
          </Pressable>
        ))}
      </View>
      <Text style={styles.note}>Карта — навигация по тренировкам, а не медицинская оценка состояния мышц.</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: {borderWidth: 1, borderRadius: 24, padding: 16, gap: 14},
  header: {gap: 12},
  eyebrow: {fontSize: 11, fontWeight: '800', letterSpacing: 1.4, opacity: 0.5},
  title: {fontSize: 22, fontWeight: '800', marginTop: 3},
  segment: {flexDirection: 'row', padding: 3, borderWidth: 1, borderRadius: 13},
  segmentItem: {flex: 1, alignItems: 'center', paddingVertical: 8, borderRadius: 10},
  segmentActive: {backgroundColor: '#111'},
  segmentText: {fontWeight: '700', fontSize: 13},
  segmentTextActive: {color: '#fff'},
  bodyCanvas: {height: 520, alignSelf: 'center', width: 230, position: 'relative'},
  head: {position: 'absolute', top: 10, left: 94, width: 42, height: 48, borderWidth: 2, borderRadius: 22, opacity: 0.28},
  neck: {position: 'absolute', top: 55, left: 103, width: 24, height: 23, borderWidth: 2, opacity: 0.22},
  torso: {position: 'absolute', top: 75, left: 60, width: 110, height: 215, borderWidth: 2, borderRadius: 42, opacity: 0.22},
  arm: {position: 'absolute', top: 92, width: 32, height: 218, borderWidth: 2, borderRadius: 20, opacity: 0.22},
  armLeft: {left: 28, transform: [{rotate: '5deg'}]},
  armRight: {right: 28, transform: [{rotate: '-5deg'}]},
  leg: {position: 'absolute', top: 279, width: 43, height: 225, borderWidth: 2, borderRadius: 22, opacity: 0.22},
  legLeft: {left: 69},
  legRight: {right: 69},
  hotspot: {position: 'absolute', zIndex: 4, minHeight: 34, borderRadius: 17, borderWidth: 1, backgroundColor: '#fff', alignItems: 'center', justifyContent: 'center', paddingHorizontal: 8},
  hotspotText: {fontSize: 12, fontWeight: '800'},
  note: {fontSize: 12, lineHeight: 17, opacity: 0.48},
});

import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, WorkoutView} from '../api/client';
import {muscleMeta, MuscleId} from '../domain/muscles';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

const recordLabels: Record<string, string> = {
  max_weight: 'Максимальный вес',
  max_reps: 'Максимум повторений',
  estimated_1rm: 'Расчётный 1ПМ',
};

export function WorkoutDetailScreen({
  accessToken,
  workoutId,
  onBack,
  onRepeat,
}: {
  accessToken: string;
  workoutId: string;
  onBack: () => void;
  onRepeat: (workout: WorkoutView) => void;
}) {
  const [workout, setWorkout] = useState<WorkoutView | null>(null);
  const [loading, setLoading] = useState(true);
  const [working, setWorking] = useState(false);
  const [error, setError] = useState('');

  async function load() {
    setLoading(true); setError('');
    try { setWorkout(await api.getWorkout(accessToken, workoutId)); }
    catch (e) { setWorkout(null); setError(e instanceof Error ? e.message : 'Не удалось загрузить тренировку.'); }
    finally { setLoading(false); }
  }

  useEffect(() => { void load(); }, [accessToken, workoutId]);

  if (loading) return <View accessibilityRole="progressbar" accessibilityLabel="Загрузка тренировки" style={styles.loading}><ActivityIndicator color={colors.primary} /><Text style={styles.muted}>Загружаем тренировку…</Text></View>;
  if (!workout) return <View style={styles.loading}><Text accessibilityRole="header" style={styles.stateTitle}>Тренировка недоступна</Text><Text accessibilityRole="alert" style={styles.muted}>{error || 'Тренировка не найдена.'}</Text><AppButton label="Повторить" variant="secondary" testID="workout-detail-retry" onPress={() => void load()} /><AppButton label="К истории" variant="text" onPress={onBack} /></View>;

  const muscle = workout.workout.muscle as MuscleId;

  async function toggleFavorite() {
    if (!workout) return;
    try {
      setWorking(true);
      setWorkout(await api.setWorkoutFavorite(accessToken, workout.workout.id, !workout.workout.favorite));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось изменить избранное.');
    } finally {
      setWorking(false);
    }
  }

  async function repeat() {
    if (!workout) return;
    try {
      setWorking(true);
      onRepeat(await api.repeatWorkout(accessToken, workout.workout.id));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось повторить тренировку.');
    } finally {
      setWorking(false);
    }
  }

  return (
    <ScrollView testID="workout-detail-screen" contentContainerStyle={styles.container}>
      <View style={styles.topline}>
        <Pressable accessibilityRole="button" accessibilityLabel="Вернуться к истории" onPress={onBack} style={styles.topAction}><Text style={styles.back}>← История</Text></Pressable>
        <Pressable accessibilityRole="button" accessibilityLabel={workout.workout.favorite?'Убрать тренировку из избранного':'Добавить тренировку в избранное'} accessibilityState={{disabled:working,busy:working}} onPress={toggleFavorite} disabled={working} style={[styles.favorite,working&&styles.disabled]}>
          <Text style={styles.favoriteText}>{workout.workout.favorite ? '★ В избранном' : '☆ В избранное'}</Text>
        </Pressable>
      </View>
      <Text style={styles.eyebrow}>ТРЕНИРОВКА</Text>
      <Text accessibilityRole="header" style={styles.title}>{muscleMeta[muscle]?.title ?? workout.workout.muscle}</Text>
      <Text style={styles.meta}>{new Date(workout.workout.completed_at ?? workout.workout.created_at).toLocaleString('ru-RU')} · {environmentLabel(workout.workout.environment)}</Text>

      <View style={styles.metrics}>
        <Metric label="Объём" value={`${Math.round(workout.workout.total_volume)} кг`} />
        <Metric label="Упражнений" value={String(workout.exercises.length)} />
        <Metric label="PR" value={String(workout.personal_records.length)} />
      </View>

      {workout.personal_records.length > 0 ? (
        <View style={styles.prCard}>
          <Text style={styles.sectionTitle}>🏆 Рекорды этой тренировки</Text>
          {workout.personal_records.map(record => {
            const exercise = workout.exercises.find(item => item.exercise.id === record.exercise_id)?.exercise;
            const suffix = record.record_type === 'max_reps' ? ' повт.' : ' кг';
            return <Text key={record.id} style={styles.prRow}>{exercise?.name ?? record.exercise_id}: {recordLabels[record.record_type]} — {record.value}{suffix}</Text>;
          })}
        </View>
      ) : null}

      <Text style={styles.sectionTitle}>Упражнения</Text>
      {workout.exercises.map(item => (
        <View key={item.workout_exercise.id} style={styles.exerciseCard}>
          <Text style={styles.exerciseTitle}>{item.exercise.name}</Text>
          <Text style={styles.target}>{item.workout_exercise.target_sets} × {item.workout_exercise.target_reps_min}–{item.workout_exercise.target_reps_max}</Text>
          {item.sets.map(set => <Text key={set.id} style={styles.set}>✓ {set.weight !== undefined ? `${set.weight} кг × ` : ''}{set.repetitions} · RIR {set.rir ?? '—'}</Text>)}
        </View>
      ))}

      {workout.exercises.length===0?<View style={styles.empty}><Text style={styles.stateTitle}>Нет записанных упражнений</Text><Text style={styles.muted}>В этой тренировке не сохранилось ни одного упражнения.</Text></View>:null}
      {error ? <Text accessibilityRole="alert" style={styles.error}>{error}</Text> : null}
      {workout.workout.status === 'completed' ? (
        <AppButton label="Повторить с прогрессией" testID="workout-detail-repeat" loading={working} onPress={() => void repeat()} />
      ) : null}
    </ScrollView>
  );
}

function Metric({label, value}: {label: string; value: string}) {
  return <View style={styles.metric}><Text style={styles.metricValue}>{value}</Text><Text style={styles.metricLabel}>{label}</Text></View>;
}

function environmentLabel(value: string) { return value==='gym'?'Зал':value==='home'?'Дома':value==='band'?'Резинки':value; }

const styles = StyleSheet.create({
  loading: {flex: 1, alignItems: 'center', justifyContent: 'center', padding: spacing.lg, gap: spacing.md, backgroundColor: colors.background},
  container: {padding: spacing.lg, paddingBottom: spacing.xl, gap: spacing.md, backgroundColor: colors.background},
  topline: {flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'},
  topAction: {minHeight: control.minTouch, justifyContent: 'center', paddingHorizontal: spacing.xs},
  back: {fontWeight: '800', color: colors.text},
  favorite: {borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, paddingHorizontal: 11, minHeight: control.minTouch, justifyContent:'center'},
  favoriteText: {fontWeight: '800', color: colors.text},
  eyebrow: {fontSize: 11, fontWeight: '800', letterSpacing: 1.4, color: colors.textMuted},
  title: {fontSize: 34, lineHeight:40, fontWeight: '900', color:colors.text},
  meta: {fontSize: 13, lineHeight:18, color:colors.textMuted},
  metrics: {flexDirection: 'row', gap: 8},
  metric: {flex: 1, borderWidth: 1, borderColor:colors.border,borderRadius: radius.md, padding: 12},
  metricValue: {fontSize: 19, fontWeight: '900',color:colors.text},
  metricLabel: {fontSize: 11, color:colors.textMuted, marginTop: 4},
  sectionTitle: {fontSize: 18, fontWeight: '900', marginTop: 4,color:colors.text},
  prCard: {borderWidth: 1,borderColor:colors.border,borderRadius: radius.lg, padding: 15, gap: 7},
  prRow: {fontSize: 13, lineHeight: 18,color:colors.text},
  exerciseCard: {borderWidth: 1,borderColor:colors.border,borderRadius: radius.lg, padding: 15, gap: 5},
  exerciseTitle: {fontSize: 17,lineHeight:23,fontWeight: '900',color:colors.text},
  target: {fontSize: 13,color:colors.textMuted},
  set: {fontSize: 13,color:colors.textMuted},
  primary: {backgroundColor: '#111', minHeight: 54, borderRadius: 16, alignItems: 'center', justifyContent: 'center', marginTop: 4},
  primaryText: {color: '#fff', fontWeight: '900'},
  error: {color: colors.danger,fontWeight:'700'}, muted:{fontSize:14,lineHeight:20,color:colors.textMuted,textAlign:'center'}, stateTitle:{fontSize:20,fontWeight:'900',color:colors.text}, empty:{borderWidth:1,borderColor:colors.border,borderRadius:radius.lg,padding:spacing.md,gap:spacing.xs}, disabled:{opacity:.45},
});

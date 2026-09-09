import React, {useEffect, useState} from 'react';
import {ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View} from 'react-native';
import {api, WorkoutView} from '../api/client';
import {muscleMeta, MuscleId} from '../domain/muscles';

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

  useEffect(() => {
    api.getWorkout(accessToken, workoutId)
      .then(setWorkout)
      .catch(e => setError(e instanceof Error ? e.message : 'Не удалось загрузить тренировку.'))
      .finally(() => setLoading(false));
  }, [accessToken, workoutId]);

  if (loading) return <View style={styles.loading}><ActivityIndicator /></View>;
  if (!workout) return <View style={styles.loading}><Text>{error || 'Тренировка не найдена.'}</Text></View>;

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
    <ScrollView contentContainerStyle={styles.container}>
      <View style={styles.topline}>
        <Pressable onPress={onBack}><Text style={styles.back}>← История</Text></Pressable>
        <Pressable onPress={toggleFavorite} disabled={working} style={styles.favorite}>
          <Text style={styles.favoriteText}>{workout.workout.favorite ? '★ В избранном' : '☆ В избранное'}</Text>
        </Pressable>
      </View>
      <Text style={styles.eyebrow}>ТРЕНИРОВКА</Text>
      <Text style={styles.title}>{muscleMeta[muscle]?.title ?? workout.workout.muscle}</Text>
      <Text style={styles.meta}>{new Date(workout.workout.completed_at ?? workout.workout.created_at).toLocaleString('ru-RU')} · {workout.workout.environment}</Text>

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

      {error ? <Text style={styles.error}>{error}</Text> : null}
      {workout.workout.status === 'completed' ? (
        <Pressable style={styles.primary} onPress={repeat} disabled={working}>
          {working ? <ActivityIndicator color="#fff" /> : <Text style={styles.primaryText}>ПОВТОРИТЬ С ПРОГРЕССИЕЙ</Text>}
        </Pressable>
      ) : null}
    </ScrollView>
  );
}

function Metric({label, value}: {label: string; value: string}) {
  return <View style={styles.metric}><Text style={styles.metricValue}>{value}</Text><Text style={styles.metricLabel}>{label}</Text></View>;
}

const styles = StyleSheet.create({
  loading: {flex: 1, alignItems: 'center', justifyContent: 'center'},
  container: {padding: 20, gap: 14},
  topline: {flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center'},
  back: {fontWeight: '800'},
  favorite: {borderWidth: 1, borderRadius: 12, paddingHorizontal: 11, paddingVertical: 8},
  favoriteText: {fontWeight: '800'},
  eyebrow: {fontSize: 11, fontWeight: '800', letterSpacing: 1.4, opacity: 0.45},
  title: {fontSize: 34, fontWeight: '900'},
  meta: {fontSize: 13, opacity: 0.55},
  metrics: {flexDirection: 'row', gap: 8},
  metric: {flex: 1, borderWidth: 1, borderRadius: 16, padding: 12},
  metricValue: {fontSize: 19, fontWeight: '900'},
  metricLabel: {fontSize: 11, opacity: 0.5, marginTop: 4},
  sectionTitle: {fontSize: 18, fontWeight: '900', marginTop: 4},
  prCard: {borderWidth: 1, borderRadius: 18, padding: 15, gap: 7},
  prRow: {fontSize: 13, lineHeight: 18},
  exerciseCard: {borderWidth: 1, borderRadius: 18, padding: 15, gap: 5},
  exerciseTitle: {fontSize: 17, fontWeight: '900'},
  target: {fontSize: 13, opacity: 0.55},
  set: {fontSize: 13, opacity: 0.72},
  primary: {backgroundColor: '#111', minHeight: 54, borderRadius: 16, alignItems: 'center', justifyContent: 'center', marginTop: 4},
  primaryText: {color: '#fff', fontWeight: '900'},
  error: {color: '#8b1e1e'},
});

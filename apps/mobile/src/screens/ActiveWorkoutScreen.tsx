import React, {useEffect, useMemo, useState} from 'react';
import {ActivityIndicator, Alert, Pressable, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, WorkoutSet, WorkoutView} from '../api/client';
import {ExerciseGuide} from '../components/ExerciseGuide';
import {sessionStorage} from '../storage/session';
import {restTimerNotifications} from '../native/restTimer';
import {TechniqueWorkoutContext, techniqueKeyForExercise} from '../domain/technique';

export function ActiveWorkoutScreen({
  accessToken,
  workout,
  onWorkoutChange,
  onFinish,
  onCancel,
  onTechnique,
  techniquePrefill,
}: {
  accessToken: string;
  workout: WorkoutView;
  onWorkoutChange: (workout: WorkoutView) => void;
  onFinish: () => Promise<void>;
  onCancel: () => Promise<void>;
  onTechnique?: (context: TechniqueWorkoutContext) => void;
  techniquePrefill?: {workoutExerciseId: string; setNumber: number; repCount: number; analysisId: string} | null;
}) {
  const [exerciseIndex, setExerciseIndex] = useState(0);
  const [weight, setWeight] = useState('');
  const [reps, setReps] = useState('');
  const [rir, setRir] = useState('2');
  const [rest, setRest] = useState(0);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [offlineNotice, setOfflineNotice] = useState('');
  const [guideOpen, setGuideOpen] = useState(false);
  const [finishing, setFinishing] = useState(false);
  const [cancelling, setCancelling] = useState(false);

  const current = workout.exercises[exerciseIndex];
  const completedSets = current?.sets.length ?? 0;
  const nextSetNumber = completedSets + 1;

  useEffect(() => {
    if (!current) return;
    setWeight(current.workout_exercise.target_weight ? String(current.workout_exercise.target_weight) : '');
    setReps(String(current.workout_exercise.target_reps_min));
  }, [current?.workout_exercise.id]);

  useEffect(() => {
    if (!current || !techniquePrefill) return;
    if (techniquePrefill.workoutExerciseId === current.workout_exercise.id && techniquePrefill.setNumber === nextSetNumber) {
      setReps(String(techniquePrefill.repCount));
      setError('');
    }
  }, [current?.workout_exercise.id, nextSetNumber, techniquePrefill?.analysisId]);

  useEffect(() => {
    if (rest <= 0) return;
    const timer = setInterval(() => setRest(value => Math.max(0, value - 1)), 1000);
    return () => clearInterval(timer);
  }, [rest]);

  const progress = useMemo(() => {
    const done = workout.exercises.reduce((sum, item) => sum + Math.min(item.sets.length, item.workout_exercise.target_sets), 0);
    const target = workout.exercises.reduce((sum, item) => sum + item.workout_exercise.target_sets, 0);
    return {done, target};
  }, [workout]);

  if (!current) return null;

  async function saveSet() {
    const repsValue = Number(reps);
    const weightValue = weight.trim() === '' ? undefined : Number(weight);
    const rirValue = rir.trim() === '' ? undefined : Number(rir);
    if (!Number.isInteger(repsValue) || repsValue < 1 || repsValue > 200) {
      setError('Повторы должны быть целым числом от 1 до 200.');
      return;
    }
    if (weightValue !== undefined && (!Number.isFinite(weightValue) || weightValue < 0 || weightValue > 1000)) {
      setError('Вес должен быть от 0 до 1000 кг.');
      return;
    }
    if (rirValue !== undefined && (!Number.isInteger(rirValue) || rirValue < 0 || rirValue > 10)) {
      setError('RIR должен быть целым числом от 0 до 10.');
      return;
    }
    const payload = {
      workout_exercise_id: current.workout_exercise.id,
      set_number: nextSetNumber,
      weight: weightValue,
      repetitions: repsValue,
      rir: rirValue,
    };
    try {
      setSaving(true);
      setError('');
      setOfflineNotice('');
      const updated = await api.logSet(accessToken, workout.workout.id, payload);
      onWorkoutChange(updated);
      startRest(current.workout_exercise.rest_seconds);
    } catch (e) {
      if (!api.isNetworkError(e)) {
        setError(e instanceof Error ? e.message : 'Не удалось сохранить подход');
        return;
      }
      const optimisticSet: WorkoutSet = {
        id: `offline-${Date.now()}`,
        workout_exercise_id: current.workout_exercise.id,
        set_number: nextSetNumber,
        weight: weightValue,
        repetitions: repsValue,
        rir: rirValue,
        completed_at: new Date().toISOString(),
      };
      const updated: WorkoutView = {
        ...workout,
        exercises: workout.exercises.map(item =>
          item.workout_exercise.id === current.workout_exercise.id
            ? {...item, sets: [...item.sets.filter(set => set.set_number !== nextSetNumber), optimisticSet].sort((a, b) => a.set_number - b.set_number)}
            : item,
        ),
      };
      await sessionStorage.enqueueSet(workout.workout.id, payload);
      await sessionStorage.saveActiveWorkout(updated);
      onWorkoutChange(updated);
      setOfflineNotice('Нет сети: подход сохранён на телефоне и будет синхронизирован автоматически.');
      startRest(current.workout_exercise.rest_seconds);
    } finally {
      setSaving(false);
    }
  }

  function startRest(seconds: number) {
    setRest(seconds);
    void restTimerNotifications.cancel().finally(() => {
      void restTimerNotifications.schedule(seconds, current.exercise.name);
    });
  }

  function skipRest() {
    setRest(0);
    void restTimerNotifications.cancel();
  }

  async function finishConfirmed() {
    try {
      setFinishing(true);
      await restTimerNotifications.cancel();
      setRest(0);
      await onFinish();
    } finally {
      setFinishing(false);
    }
  }

  function requestFinish() {
    if (progress.done < progress.target) {
      Alert.alert(
        'Завершить раньше?',
        `Выполнено ${progress.done} из ${progress.target} запланированных подходов. Незавершённые подходы останутся пропущенными.`,
        [
          {text: 'Продолжить тренировку', style: 'cancel'},
          {text: 'Завершить', style: 'destructive', onPress: () => { void finishConfirmed(); }},
        ],
      );
      return;
    }
    void finishConfirmed();
  }

  function requestCancel() {
    Alert.alert(
      'Отменить тренировку?',
      'Записанные подходы останутся в отменённой тренировке, но не будут учитываться как завершённая тренировка.',
      [
        {text: 'Остаться', style: 'cancel'},
        {
          text: 'Отменить тренировку',
          style: 'destructive',
          onPress: () => {
            setCancelling(true);
            void restTimerNotifications.cancel().finally(() => {
              setRest(0);
              void onCancel().finally(() => setCancelling(false));
            });
          },
        },
      ],
    );
  }

  return (
    <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
      <Text style={styles.eyebrow}>АКТИВНАЯ ТРЕНИРОВКА</Text>
      <Text style={styles.progress}>{progress.done} / {progress.target} подходов</Text>

      <View style={styles.card}>
        <Text style={styles.index}>{exerciseIndex + 1} / {workout.exercises.length}</Text>
        <Text style={styles.title}>{current.exercise.name}</Text>
        <Text style={styles.target}>
          Цель: {current.workout_exercise.target_sets} × {current.workout_exercise.target_reps_min}–{current.workout_exercise.target_reps_max}
          {current.workout_exercise.target_weight ? ` · ${current.workout_exercise.target_weight} кг` : ''}
        </Text>
        <Pressable style={styles.technique} onPress={() => setGuideOpen(true)}>
          <Text style={styles.techniqueText}>▶ Показать технику</Text>
        </Pressable>
        {onTechnique && nextSetNumber <= current.workout_exercise.target_sets && techniqueKeyForExercise(current.exercise.id) ? <Pressable
          style={styles.liveTechnique}
          onPress={() => onTechnique({
            workoutId: workout.workout.id,
            workoutExerciseId: current.workout_exercise.id,
            setNumber: nextSetNumber,
            exerciseId: current.exercise.id,
            exerciseName: current.exercise.name,
          })}
          accessibilityRole="button"
          accessibilityLabel="Открыть live-анализ техники"
          testID="workout-live-technique">
          <Text style={styles.liveTechniqueText}>● LIVE · считать повторы камерой</Text>
        </Pressable> : null}

        <View style={styles.inputs}>
          <View style={styles.field}>
            <Text style={styles.label}>Вес, кг</Text>
            <TextInput value={weight} onChangeText={setWeight} keyboardType="decimal-pad" style={styles.input} placeholder="—" />
          </View>
          <View style={styles.field}>
            <Text style={styles.label}>Повторы</Text>
            <TextInput value={reps} onChangeText={setReps} keyboardType="number-pad" style={styles.input} />
          </View>
          <View style={styles.field}>
            <Text style={styles.label}>RIR</Text>
            <TextInput value={rir} onChangeText={setRir} keyboardType="decimal-pad" style={styles.input} />
          </View>
        </View>

        <Text style={styles.setTitle}>Подход {nextSetNumber}</Text>
        {current.sets.map(set => (
          <Text key={set.id} style={styles.savedSet}>
            ✓ {set.set_number}. {set.weight !== undefined ? `${set.weight} кг × ` : ''}{set.repetitions} · RIR {set.rir ?? '—'}
          </Text>
        ))}

        {error ? <Text style={styles.error}>{error}</Text> : null}
        {offlineNotice ? <Text style={styles.offline}>{offlineNotice}</Text> : null}

        <Pressable style={[styles.primary, saving && styles.disabled]} onPress={saveSet} disabled={saving} accessibilityRole="button" accessibilityLabel="Завершить подход" testID="workout-save-set">
          {saving ? <ActivityIndicator color="#fff" /> : <Text style={styles.primaryText}>ЗАВЕРШИТЬ ПОДХОД</Text>}
        </Pressable>
      </View>

      {rest > 0 ? (
        <View style={styles.restCard}>
          <Text style={styles.restLabel}>Отдых</Text>
          <Text style={styles.restValue}>{Math.floor(rest / 60)}:{String(rest % 60).padStart(2, '0')}</Text>
          <Pressable onPress={skipRest}><Text style={styles.skip}>Пропустить таймер</Text></Pressable>
        </View>
      ) : null}

      <View style={styles.navigation}>
        <Pressable disabled={exerciseIndex === 0} onPress={() => setExerciseIndex(index => Math.max(0, index - 1))} style={styles.secondary}>
          <Text style={styles.secondaryText}>← Предыдущее</Text>
        </Pressable>
        <Pressable disabled={exerciseIndex === workout.exercises.length - 1} onPress={() => setExerciseIndex(index => Math.min(workout.exercises.length - 1, index + 1))} style={styles.secondary}>
          <Text style={styles.secondaryText}>Следующее →</Text>
        </Pressable>
      </View>

      <Pressable style={[styles.finish, finishing && styles.disabled]} onPress={requestFinish} disabled={finishing || cancelling} accessibilityRole="button" accessibilityLabel="Завершить тренировку" testID="workout-finish">
        {finishing ? <ActivityIndicator /> : <Text style={styles.finishText}>Завершить тренировку</Text>}
      </Pressable>
      <Pressable style={[styles.cancel, cancelling && styles.disabled]} onPress={requestCancel} disabled={finishing || cancelling} accessibilityRole="button" accessibilityLabel="Отменить тренировку" testID="workout-cancel">
        {cancelling ? <ActivityIndicator /> : <Text style={styles.cancelText}>Отменить тренировку</Text>}
      </Pressable>
      <ExerciseGuide exercise={current.exercise} visible={guideOpen} onClose={() => setGuideOpen(false)} />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {padding: 20, gap: 14},
  eyebrow: {fontSize: 12, letterSpacing: 1.4, fontWeight: '800'},
  progress: {fontSize: 14, fontWeight: '700', opacity: 0.55},
  card: {borderWidth: 1, borderRadius: 22, padding: 18, gap: 12},
  index: {fontSize: 13, fontWeight: '700', opacity: 0.45},
  title: {fontSize: 27, lineHeight: 32, fontWeight: '800'},
  target: {fontSize: 15, lineHeight: 21, opacity: 0.65},
  technique: {alignSelf: 'flex-start', borderWidth: 1, borderRadius: 12, paddingVertical: 9, paddingHorizontal: 12},
  liveTechnique: {alignSelf: 'flex-start', backgroundColor: '#111', borderRadius: 12, paddingVertical: 10, paddingHorizontal: 13},
  liveTechniqueText: {color: '#fff', fontWeight: '900'},
  techniqueText: {fontWeight: '800'},
  inputs: {flexDirection: 'row', gap: 8},
  field: {flex: 1, gap: 6},
  label: {fontSize: 12, fontWeight: '700', opacity: 0.55},
  input: {borderWidth: 1, borderRadius: 12, paddingHorizontal: 12, paddingVertical: 12, fontSize: 17, fontWeight: '700'},
  setTitle: {fontSize: 15, fontWeight: '800', marginTop: 4},
  savedSet: {fontSize: 14, opacity: 0.72},
  primary: {backgroundColor: '#111', paddingVertical: 16, borderRadius: 15, alignItems: 'center'},
  primaryText: {color: '#fff', fontWeight: '800'},
  error: {color: '#8b1e1e'},
  offline: {fontSize: 13, lineHeight: 18, fontWeight: '650', opacity: 0.7},
  restCard: {borderWidth: 1, borderRadius: 20, padding: 18, alignItems: 'center', gap: 4},
  restLabel: {fontWeight: '700', opacity: 0.55},
  restValue: {fontSize: 42, fontWeight: '800'},
  skip: {fontWeight: '700', marginTop: 4},
  navigation: {flexDirection: 'row', gap: 10},
  secondary: {flex: 1, paddingVertical: 14, borderWidth: 1, borderRadius: 14, alignItems: 'center'},
  secondaryText: {fontWeight: '700'},
  finish: {paddingVertical: 15, alignItems: 'center'},
  finishText: {fontWeight: '800', opacity: 0.65},
  cancel: {minHeight: 48, paddingVertical: 14, alignItems: 'center', justifyContent: 'center', borderWidth: 1, borderRadius: 14},
  cancelText: {fontWeight: '800', color: '#8b1e1e'},
  disabled: {opacity: 0.5},
});

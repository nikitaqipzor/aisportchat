import React, {useEffect, useMemo, useState} from 'react';
import {ActivityIndicator, Alert, AppState, Pressable, ScrollView, StyleSheet, Text, TextInput, View} from 'react-native';
import {api, WorkoutSet, WorkoutView} from '../api/client';
import {ExerciseGuide} from '../components/ExerciseGuide';
import {sessionStorage} from '../storage/session';
import {restSecondsRemaining, restTimerCoordinator} from '../domain/restTimer';
import {TechniqueWorkoutContext, techniqueKeyForExercise} from '../domain/technique';
import {AppButton} from '../components/AppButton';
import {colors, control, radius, spacing} from '../theme/tokens';

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
  const [restDeadline, setRestDeadline] = useState<number | null>(null);
  const [clock, setClock] = useState(() => Date.now());
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

  const rest = restSecondsRemaining(restDeadline, clock);

  useEffect(() => {
    let active = true;
    void restTimerCoordinator.restore(workout.workout.id).then(snapshot => {
      if (active) {
        setClock(Date.now());
        setRestDeadline(snapshot?.deadlineMs ?? null);
      }
    });
    const subscription = AppState.addEventListener('change', state => {
      if (state !== 'active') return;
      setClock(Date.now());
      void restTimerCoordinator.restore(workout.workout.id).then(snapshot => {
        if (active) setRestDeadline(snapshot?.deadlineMs ?? null);
      });
    });
    return () => {
      active = false;
      subscription.remove();
    };
  }, [workout.workout.id]);

  useEffect(() => {
    if (restDeadline === null) return;
    const tick = () => setClock(Date.now());
    tick();
    const timer = setInterval(tick, 1000);
    return () => clearInterval(timer);
  }, [restDeadline]);

  useEffect(() => {
    if (restDeadline !== null && rest <= 0) {
      setRestDeadline(null);
      void restTimerCoordinator.invalidate();
    }
  }, [rest, restDeadline]);

  const progress = useMemo(() => {
    const done = workout.exercises.reduce((sum, item) => sum + Math.min(item.sets.length, item.workout_exercise.target_sets), 0);
    const target = workout.exercises.reduce((sum, item) => sum + item.workout_exercise.target_sets, 0);
    return {done, target};
  }, [workout]);

  if (!current) return <View style={styles.emptyState}><Text accessibilityRole="header" style={styles.emptyTitle}>В тренировке нет упражнений</Text><Text style={styles.emptyText}>Вернись назад и создай тренировку ещё раз.</Text><AppButton label="Отменить тренировку" variant="secondary" onPress={() => { void onCancel(); }} /></View>;

  async function saveSet() {
    if (nextSetNumber > current.workout_exercise.target_sets) {
      setError('Все запланированные подходы упражнения уже выполнены. Перейди к следующему.');
      return;
    }
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
    const deadline = Date.now() + seconds * 1000;
    setClock(Date.now());
    setRestDeadline(deadline);
    void restTimerCoordinator.start(workout.workout.id, current.exercise.name, seconds);
  }

  function skipRest() {
    setRestDeadline(null);
    void restTimerCoordinator.invalidate();
  }

  async function finishConfirmed() {
    try {
      setFinishing(true);
      setError('');
      setRestDeadline(null);
      await restTimerCoordinator.invalidate();
      await onFinish();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось завершить тренировку.');
    } finally {
      setFinishing(false);
    }
  }

  function requestFinish() {
    const early = progress.done < progress.target;
    Alert.alert(
      early ? 'Завершить раньше?' : 'Завершить тренировку?',
      early
        ? `Выполнено ${progress.done} из ${progress.target} запланированных подходов. Незавершённые подходы останутся пропущенными.`
        : `Все ${progress.target} подходов выполнены. Сохранить тренировку в истории?`,
      [
        {text: early ? 'Продолжить тренировку' : 'Отмена', style: 'cancel'},
        {text: 'Завершить', onPress: () => { void finishConfirmed(); }},
      ],
    );
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
            setRestDeadline(null);
            void restTimerCoordinator.invalidate().then(onCancel).catch(e => setError(e instanceof Error ? e.message : 'Не удалось отменить тренировку.')).finally(() => setCancelling(false));
          },
        },
      ],
    );
  }

  return (
    <ScrollView testID="active-workout-screen" contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
      <Text style={styles.eyebrow}>АКТИВНАЯ ТРЕНИРОВКА</Text>
      <Text accessibilityRole="header" style={styles.progress}>{progress.done} из {progress.target} подходов</Text>
      <View accessibilityRole="progressbar" accessibilityValue={{min:0,max:progress.target||1,now:progress.done}} style={styles.progressTrack}><View style={[styles.progressFill,{width:`${progress.target?Math.min(100,progress.done/progress.target*100):0}%`}]} /></View>

      <View style={styles.card}>
        <Text style={styles.index}>{exerciseIndex + 1} / {workout.exercises.length}</Text>
        <Text style={styles.title}>{current.exercise.name}</Text>
        <Text style={styles.target}>
          Цель: {current.workout_exercise.target_sets} × {current.workout_exercise.target_reps_min}–{current.workout_exercise.target_reps_max}
          {current.workout_exercise.target_weight ? ` · ${current.workout_exercise.target_weight} кг` : ''}
        </Text>
        <Pressable accessibilityRole="button" accessibilityLabel={`Показать технику упражнения ${current.exercise.name}`} style={({pressed})=>[styles.technique,pressed&&styles.pressed]} onPress={() => setGuideOpen(true)}>
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
            <TextInput testID="workout-weight" accessibilityLabel="Вес в килограммах" value={weight} onChangeText={setWeight} keyboardType="decimal-pad" style={styles.input} placeholder="—" placeholderTextColor={colors.textMuted} />
          </View>
          <View style={styles.field}>
            <Text style={styles.label}>Повторы</Text>
            <TextInput testID="workout-reps" accessibilityLabel="Количество повторений" value={reps} onChangeText={setReps} keyboardType="number-pad" style={styles.input} />
          </View>
          <View style={styles.field}>
            <Text style={styles.label}>RIR</Text>
            <TextInput testID="workout-rir" accessibilityLabel="Повторения в запасе, RIR" value={rir} onChangeText={setRir} keyboardType="number-pad" style={styles.input} />
          </View>
        </View>

        <Text style={styles.setTitle}>Подход {nextSetNumber}</Text>
        {current.sets.map(set => (
          <Text key={set.id} style={styles.savedSet}>
            ✓ {set.set_number}. {set.weight !== undefined ? `${set.weight} кг × ` : ''}{set.repetitions} · RIR {set.rir ?? '—'}
          </Text>
        ))}

        {error ? <Text accessibilityRole="alert" style={styles.error}>{error}</Text> : null}
        {offlineNotice ? <Text accessibilityRole="alert" style={styles.offline}>{offlineNotice}</Text> : null}

        <Pressable style={[styles.primary, (saving||finishing||cancelling||nextSetNumber>current.workout_exercise.target_sets) && styles.disabled]} onPress={()=>void saveSet()} disabled={saving||finishing||cancelling||nextSetNumber>current.workout_exercise.target_sets} accessibilityRole="button" accessibilityLabel="Завершить подход" accessibilityState={{disabled:saving||finishing||cancelling||nextSetNumber>current.workout_exercise.target_sets,busy:saving}} testID="workout-save-set">
          {saving ? <ActivityIndicator color={colors.inverse} /> : <Text style={styles.primaryText}>{nextSetNumber>current.workout_exercise.target_sets?'Все подходы выполнены':'Завершить подход'}</Text>}
        </Pressable>
      </View>

      {rest > 0 ? (
        <View style={styles.restCard}>
          <Text style={styles.restLabel}>Отдых</Text>
          <Text accessibilityLiveRegion="polite" accessibilityLabel={`Осталось отдыха ${Math.floor(rest / 60)} минут ${rest % 60} секунд`} style={styles.restValue}>{Math.floor(rest / 60)}:{String(rest % 60).padStart(2, '0')}</Text>
          <Pressable accessibilityRole="button" accessibilityLabel="Пропустить таймер отдыха" onPress={skipRest} style={styles.textAction}><Text style={styles.skip}>Пропустить таймер</Text></Pressable>
        </View>
      ) : null}

      <View style={styles.navigation}>
        <Pressable accessibilityRole="button" accessibilityLabel="Предыдущее упражнение" accessibilityState={{disabled:exerciseIndex===0}} disabled={exerciseIndex === 0} onPress={() => setExerciseIndex(index => Math.max(0, index - 1))} style={[styles.secondary,exerciseIndex===0&&styles.disabled]}>
          <Text style={styles.secondaryText}>← Предыдущее</Text>
        </Pressable>
        <Pressable accessibilityRole="button" accessibilityLabel="Следующее упражнение" accessibilityState={{disabled:exerciseIndex===workout.exercises.length-1}} disabled={exerciseIndex === workout.exercises.length - 1} onPress={() => setExerciseIndex(index => Math.min(workout.exercises.length - 1, index + 1))} style={[styles.secondary,exerciseIndex===workout.exercises.length-1&&styles.disabled]}>
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
  container: {padding: spacing.lg,paddingBottom:spacing.xl,gap: spacing.md,backgroundColor:colors.background},
  eyebrow: {fontSize: 12, letterSpacing: 1.4, fontWeight: '800',color:colors.textMuted},
  progress: {fontSize: 16, fontWeight: '800',color:colors.text},progressTrack:{height:6,borderRadius:radius.pill,backgroundColor:colors.surfaceMuted,overflow:'hidden'},progressFill:{height:6,borderRadius:radius.pill,backgroundColor:colors.primary},
  card: {borderWidth: 1,borderColor:colors.border,borderRadius: radius.lg, padding: 18, gap: 12},
  index: {fontSize: 13, fontWeight: '700',color:colors.textMuted},
  title: {fontSize: 27, lineHeight: 32, fontWeight: '800',color:colors.text},
  target: {fontSize: 15, lineHeight: 21,color:colors.textMuted},
  technique: {alignSelf: 'flex-start', borderWidth: 1,borderColor:colors.border,borderRadius: radius.md,minHeight:control.minTouch,justifyContent:'center',paddingHorizontal: 12},
  liveTechnique: {alignSelf: 'flex-start', backgroundColor: colors.primary,borderRadius: radius.md,minHeight:control.minTouch,justifyContent:'center',paddingHorizontal: 13},
  liveTechniqueText: {color: colors.inverse, fontWeight: '900'},
  techniqueText: {fontWeight: '800',color:colors.text},
  inputs: {flexDirection: 'row', gap: 8},
  field: {flex: 1, gap: 6},
  label: {fontSize: 12, fontWeight: '700',color:colors.textMuted},
  input: {minHeight:control.minTouch,borderWidth: 1,borderColor:colors.border,borderRadius: radius.md,paddingHorizontal: 12,fontSize: 17, fontWeight: '700',color:colors.text},
  setTitle: {fontSize: 15, fontWeight: '800', marginTop: 4,color:colors.text},
  savedSet: {fontSize: 14,color:colors.textMuted},
  primary: {backgroundColor: colors.primary,minHeight:control.buttonHeight,borderRadius: radius.md,alignItems:'center',justifyContent:'center'},
  primaryText: {color: colors.inverse, fontWeight: '800'},
  error: {color: colors.danger,fontWeight:'700'},
  offline: {fontSize: 13, lineHeight: 18, fontWeight: '600',color:colors.success},
  restCard: {borderWidth: 1,borderColor:colors.border,borderRadius: radius.lg, padding: 18, alignItems: 'center', gap: 4},
  restLabel: {fontWeight: '700',color:colors.textMuted},
  restValue: {fontSize: 42, fontWeight: '800',color:colors.text},
  textAction:{minHeight:control.minTouch,justifyContent:'center'},skip: {fontWeight: '700',color:colors.text},
  navigation: {flexDirection: 'row', gap: 10},
  secondary: {flex: 1,minHeight:control.minTouch,borderWidth: 1,borderColor:colors.border,borderRadius: radius.md,alignItems:'center',justifyContent:'center'},
  secondaryText: {fontWeight: '700',color:colors.text},
  finish: {minHeight:control.minTouch,alignItems:'center',justifyContent:'center'},
  finishText: {fontWeight: '800',color:colors.textMuted},
  cancel: {minHeight: control.minTouch,alignItems:'center',justifyContent:'center',borderWidth:1,borderColor:colors.danger,borderRadius:radius.md},
  cancelText: {fontWeight: '800', color: colors.danger},
  disabled: {opacity: 0.5},
  pressed:{opacity:.65},emptyState:{flex:1,justifyContent:'center',padding:spacing.lg,gap:spacing.md,backgroundColor:colors.background},emptyTitle:{fontSize:22,fontWeight:'900',color:colors.text},emptyText:{fontSize:14,lineHeight:20,color:colors.textMuted},
});

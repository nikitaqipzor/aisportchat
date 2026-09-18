import type {TechniqueExercise} from '../api/client';

export type TechniqueKey = TechniqueExercise['key'];
export type TechniqueWorkoutContext = {
  workoutId: string;
  workoutExerciseId: string;
  setNumber: number;
  exerciseId: string;
  exerciseName: string;
};

export function techniqueKeyForExercise(exerciseId: string): TechniqueKey | null {
  const id = exerciseId.toLowerCase().trim().replace(/[ -]+/g, '_');
  if (id === 'pushup' || id.includes('push_up') || id.includes('pushup')) return 'push_up';
  if (id.includes('split_squat') || id.includes('lunge')) return 'lunge';
  if (id.includes('squat')) return 'squat';
  if (id.includes('biceps_curl') || id.includes('bicep_curl') || id.includes('dumbbell_curl') || id.includes('barbell_curl') || id.includes('hammer_curl')) return 'biceps_curl';
  if (id.includes('overhead_press') || id.includes('shoulder_press') || id.includes('military_press') || id.includes('arnold_press')) return 'shoulder_press';
  return null;
}

export function techniqueCaptureError(error: unknown): {cancelled: boolean; message: string} {
  const value = error as {code?: string; message?: string} | null;
  const code = value?.code ?? '';
  const message = value?.message ?? '';
  const cancelled = code.includes('CANCELLED') || /(?:отмен[аеё]н|cancelled)/i.test(message);
  if (cancelled) return {cancelled: true, message: 'Анализ отменён. Результат не сохранён.'};
  if (code.includes('PERMISSION') || /доступ к камере|camera permission/i.test(message)) {
    return {cancelled: false, message: 'Нет доступа к камере. Разрешите камеру для Athletica AI в настройках Android и попробуйте снова.'};
  }
  return {cancelled: false, message: message || 'Не удалось выполнить анализ техники'};
}

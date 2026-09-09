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
  const id = exerciseId.toLowerCase();
  if (id === 'pushup' || id.includes('push_up')) return 'push_up';
  if (id.includes('split_squat') || id.includes('lunge')) return 'lunge';
  if (id.includes('squat')) return 'squat';
  if (id.includes('curl')) return 'biceps_curl';
  if (id.includes('overhead_press') || id.includes('shoulder_press')) return 'shoulder_press';
  return null;
}

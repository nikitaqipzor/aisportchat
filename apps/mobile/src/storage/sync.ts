import {api} from '../api/client';
import type {AuthTokens, FinishResult, WorkoutView} from '../api/client';
import {sessionStorage} from './session';
import type {OfflineOperation} from './session';

export type SyncResult = {
  tokens: AuthTokens;
  workout?: WorkoutView;
  finish?: FinishResult;
  remaining: number;
};

export async function flushOfflineQueue(tokens: AuthTokens): Promise<SyncResult> {
  const currentTokens = tokens;
  const ownerAtStart = await sessionStorage.currentUserId();
  const ownerIsCurrent = async () => ownerAtStart !== undefined && (await sessionStorage.currentUserId()) === ownerAtStart;
  const queue = await sessionStorage.loadQueue();
  const remaining: OfflineOperation[] = [];
  let latestWorkout: WorkoutView | undefined;
  let finish: FinishResult | undefined;

  for (let index = 0; index < queue.length; index += 1) {
    if (!(await ownerIsCurrent())) return {tokens: currentTokens, remaining: queue.length};
    const operation = queue[index];
    try {
      if (operation.type === 'log_set') {
        latestWorkout = await api.logSet(currentTokens.access_token, operation.workoutId, operation.payload);
      } else if (operation.type === 'finish_workout') {
        finish = await api.finishWorkout(currentTokens.access_token, operation.workoutId);
        latestWorkout = finish.workout;
      } else {
        latestWorkout = await api.cancelWorkout(currentTokens.access_token, operation.workoutId);
      }
      if (!(await ownerIsCurrent())) return {tokens: currentTokens, remaining: queue.length};
    } catch (error) {
      remaining.push(...queue.slice(index));
      break;
    }
  }

  if (!(await ownerIsCurrent())) return {tokens: currentTokens, remaining: queue.length};
  await sessionStorage.saveQueue(remaining);
  if (latestWorkout) {
    await sessionStorage.saveActiveWorkout(latestWorkout.workout.status === 'active' ? latestWorkout : null);
  }
  return {tokens: currentTokens, workout: latestWorkout, finish, remaining: remaining.length};
}

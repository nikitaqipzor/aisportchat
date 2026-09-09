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
  let currentTokens = tokens;
  const queue = await sessionStorage.loadQueue();
  const remaining: OfflineOperation[] = [];
  let latestWorkout: WorkoutView | undefined;
  let finish: FinishResult | undefined;

  for (let index = 0; index < queue.length; index += 1) {
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
    } catch (error) {
      // Central API auth handles normal 401 refresh. Keep this fallback so the
      // queue still works in isolated tests or before the adapter is configured.
      if (api.isUnauthorized(error)) {
        const refreshed = await api.refresh(currentTokens.refresh_token);
        currentTokens = refreshed.tokens;
        api.setCurrentTokens(currentTokens);
        await sessionStorage.saveTokens(currentTokens);
        index -= 1;
        continue;
      }
      remaining.push(...queue.slice(index));
      break;
    }
  }

  await sessionStorage.saveQueue(remaining);
  if (latestWorkout) {
    await sessionStorage.saveActiveWorkout(latestWorkout.workout.status === 'active' ? latestWorkout : null);
  }
  return {tokens: currentTokens, workout: latestWorkout, finish, remaining: remaining.length};
}

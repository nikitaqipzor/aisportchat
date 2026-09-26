import {api} from '../api/client';
import type {AuthTokens, FinishResult, WorkoutView} from '../api/client';
import {sessionStorage} from './session';

export type SyncResult = {
  tokens: AuthTokens;
  workout?: WorkoutView;
  finish?: FinishResult;
  remaining: number;
};

// App resume and screen actions may start sync at the same time.
let previousFlush: Promise<void> = Promise.resolve();

export function flushOfflineQueue(tokens: AuthTokens): Promise<SyncResult> {
  const result = previousFlush.then(() => flushOnce(tokens));
  previousFlush = result.then(() => undefined, () => undefined);
  return result;
}

async function flushOnce(tokens: AuthTokens): Promise<SyncResult> {
  const currentTokens = tokens;
  const ownerAtStart = await sessionStorage.currentUserId();
  const ownerIsCurrent = async () => ownerAtStart !== undefined && (await sessionStorage.currentUserId()) === ownerAtStart;
  const queue = await sessionStorage.loadQueue();
  let latestWorkout: WorkoutView | undefined;
  let finish: FinishResult | undefined;

  for (const operation of queue) {
    if (!(await ownerIsCurrent())) return {tokens: currentTokens, remaining: queue.length};
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
      // Delete only the acknowledged ID. A newer enqueue of this same set has a
      // different ID and must survive even if it landed during the network call.
      const remaining = await sessionStorage.removeProcessedQueueOperations(ownerAtStart!, [operation.id]);
      if (remaining === null) return {tokens: currentTokens, remaining: queue.length};
    } catch {
      break;
    }
  }

  if (!(await ownerIsCurrent())) return {tokens: currentTokens, remaining: queue.length};
  // A server snapshot taken before a concurrent offline set must not replace
  // its optimistic local workout (or be sent to App as an older workout).
  if (latestWorkout && !await sessionStorage.saveSyncedWorkoutIfNoPendingSet(ownerAtStart!, latestWorkout)) {
    latestWorkout = undefined;
  }
  const pending = await sessionStorage.loadQueue();
  return {tokens: currentTokens, workout: latestWorkout, finish, remaining: pending.length};
}

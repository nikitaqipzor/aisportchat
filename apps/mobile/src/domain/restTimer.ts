import AsyncStorage from '@react-native-async-storage/async-storage';
import {restTimerNotifications} from '../native/restTimer';
import {currentOwnerUserId} from '../storage/session';

const STORAGE_KEY = 'fitness.rest-timer.v1';

export type RestTimerSnapshot = {
  ownerUserId: string;
  workoutId: string;
  exerciseName: string;
  deadlineMs: number;
};

type RestTimerDependencies = {
  now: () => number;
  owner: () => Promise<string | undefined>;
  read: () => Promise<string | null>;
  write: (value: string) => Promise<void>;
  remove: () => Promise<void>;
  schedule: (seconds: number, exerciseName: string) => Promise<boolean>;
  cancel: () => Promise<boolean>;
};

export function restSecondsRemaining(deadlineMs: number | null, nowMs = Date.now()) {
  return deadlineMs === null ? 0 : Math.max(0, Math.ceil((deadlineMs - nowMs) / 1000));
}

export class RestTimerCoordinator {
  private generation = 0;
  private tail: Promise<unknown> = Promise.resolve();
  private memory: RestTimerSnapshot | null = null;

  constructor(private readonly dependencies: RestTimerDependencies) {}

  private enqueue<T>(task: () => Promise<T>): Promise<T> {
    const result = this.tail.catch(() => undefined).then(task);
    this.tail = result.catch(() => undefined);
    return result;
  }

  private async removePersisted() {
    this.memory = null;
    try {
      await this.dependencies.remove();
    } catch {
      // AsyncStorage is an optimisation: the native alarm and in-memory state still work.
    }
  }

  private async compensate() {
    await this.removePersisted();
    try {
      await this.dependencies.cancel();
    } catch {
      // Native cancellation is best effort and must not poison the FIFO.
    }
  }

  start(workoutId: string, exerciseName: string, seconds: number) {
    const token = this.generation;
    const deadlineMs = this.dependencies.now() + Math.max(0, seconds) * 1000;
    return this.enqueue(async () => {
      let ownerUserId: string | undefined;
      try {
        ownerUserId = await this.dependencies.owner();
      } catch {
        await this.compensate();
        return null;
      }
      if (token !== this.generation || !ownerUserId || seconds <= 0) {
        await this.compensate();
        return null;
      }
      const snapshot: RestTimerSnapshot = {ownerUserId, workoutId, exerciseName, deadlineMs};
      this.memory = snapshot;
      try {
        await this.dependencies.write(JSON.stringify(snapshot));
      } catch {
        // Continue in memory and with the native alarm when storage is unavailable.
      }
      if (token !== this.generation) {
        await this.compensate();
        return null;
      }
      try {
        await this.dependencies.cancel();
      } catch {
        // Scheduling below replaces the alarm on Android even if cancellation failed.
      }
      if (token !== this.generation) {
        await this.compensate();
        return null;
      }
      const remaining = restSecondsRemaining(deadlineMs, this.dependencies.now());
      if (remaining <= 0) {
        await this.compensate();
        return null;
      }
      try {
        await this.dependencies.schedule(remaining, exerciseName);
      } catch {
        // The visible timer remains useful when notifications are unavailable.
      }
      if (token !== this.generation) {
        await this.compensate();
        return null;
      }
      return snapshot;
    });
  }

  restore(workoutId: string) {
    const token = this.generation;
    return this.enqueue(async () => {
      let ownerUserId: string | undefined;
      try {
        ownerUserId = await this.dependencies.owner();
      } catch {
        return null;
      }
      if (token !== this.generation || !ownerUserId) return null;
      let snapshot = this.memory;
      try {
        const raw = await this.dependencies.read();
        if (raw) snapshot = JSON.parse(raw) as RestTimerSnapshot;
      } catch {
        // Fall back to the last in-memory snapshot on corrupt/unavailable storage.
      }
      if (
        token !== this.generation ||
        !snapshot ||
        snapshot.ownerUserId !== ownerUserId ||
        snapshot.workoutId !== workoutId ||
        restSecondsRemaining(snapshot.deadlineMs, this.dependencies.now()) <= 0
      ) {
        await this.compensate();
        return null;
      }
      this.memory = snapshot;
      return snapshot;
    });
  }

  invalidate() {
    this.generation += 1;
    this.memory = null;
    return this.enqueue(async () => {
      await this.removePersisted();
      try {
        await this.dependencies.cancel();
      } catch {
        // A failed native task must not block the next timer start.
      }
    });
  }
}

export const restTimerCoordinator = new RestTimerCoordinator({
  now: Date.now,
  owner: currentOwnerUserId,
  read: () => AsyncStorage.getItem(STORAGE_KEY),
  write: value => AsyncStorage.setItem(STORAGE_KEY, value),
  remove: () => AsyncStorage.removeItem(STORAGE_KEY),
  schedule: (seconds, exerciseName) => restTimerNotifications.schedule(seconds, exerciseName),
  cancel: () => restTimerNotifications.cancel(),
});

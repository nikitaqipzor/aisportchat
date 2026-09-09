import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Keychain from 'react-native-keychain';
import type {AIChatMessage, AuthTokens, WorkoutView} from '../api/client';

const TOKEN_SERVICE = 'ai-fitness-os.auth';
const LEGACY_USERNAME = 'fitness-user';
const ACTIVE_WORKOUT_KEY = 'fitness.active-workout.v2';
const OFFLINE_QUEUE_KEY = 'fitness.offline-queue.v2';
const AI_CHAT_KEY = 'fitness.ai-chat.v2';
const LEGACY_TRANSIENT_KEYS = ['fitness.active-workout.v1', 'fitness.offline-queue.v1', 'fitness.ai-chat.v1'];

export type OfflineOperation =
  | {
      id: string;
      type: 'log_set';
      workoutId: string;
      payload: {
        workout_exercise_id: string;
        set_number: number;
        weight?: number;
        repetitions: number;
        rpe?: number;
        rir?: number;
      };
      createdAt: string;
    }
  | {
      id: string;
      type: 'finish_workout' | 'cancel_workout';
      workoutId: string;
      createdAt: string;
    };

type OwnedValue<T> = {ownerUserId: string; value: T};
export type StoredSession = {tokens: AuthTokens; userId?: string};

function newLocalId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export async function currentOwnerUserId(): Promise<string | undefined> {
  const credentials = await Keychain.getGenericPassword({service: TOKEN_SERVICE});
  if (!credentials || credentials.username === LEGACY_USERNAME) return undefined;
  return credentials.username;
}

async function saveOwned<T>(key: string, value: T | null) {
  if (value === null) {
    await AsyncStorage.removeItem(key);
    return;
  }
  const ownerUserId = await currentOwnerUserId();
  if (!ownerUserId) {
    // Never persist user data if it cannot be scoped to an authenticated account.
    await AsyncStorage.removeItem(key);
    return;
  }
  const envelope: OwnedValue<T> = {ownerUserId, value};
  await AsyncStorage.setItem(key, JSON.stringify(envelope));
}

async function loadOwned<T>(key: string): Promise<T | null> {
  const raw = await AsyncStorage.getItem(key);
  if (!raw) return null;
  const ownerUserId = await currentOwnerUserId();
  if (!ownerUserId) {
    await AsyncStorage.removeItem(key);
    return null;
  }
  try {
    const envelope = JSON.parse(raw) as Partial<OwnedValue<T>>;
    if (!envelope.ownerUserId || envelope.ownerUserId !== ownerUserId || envelope.value === undefined) {
      await AsyncStorage.removeItem(key);
      return null;
    }
    return envelope.value as T;
  } catch {
    await AsyncStorage.removeItem(key);
    return null;
  }
}

export const sessionStorage = {
  currentUserId: currentOwnerUserId,
  async saveTokens(tokens: AuthTokens, userId?: string) {
    const existing = await Keychain.getGenericPassword({service: TOKEN_SERVICE});
    const currentUser = existing && existing.username !== LEGACY_USERNAME ? existing.username : undefined;
    const owner = userId ?? currentUser ?? LEGACY_USERNAME;
    await Keychain.setGenericPassword(owner, JSON.stringify(tokens), {
      service: TOKEN_SERVICE,
      accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
    });
  },

  async loadSession(): Promise<StoredSession | null> {
    const credentials = await Keychain.getGenericPassword({service: TOKEN_SERVICE});
    if (!credentials) return null;
    try {
      return {
        tokens: JSON.parse(credentials.password) as AuthTokens,
        userId: credentials.username === LEGACY_USERNAME ? undefined : credentials.username,
      };
    } catch {
      await Keychain.resetGenericPassword({service: TOKEN_SERVICE});
      return null;
    }
  },

  async loadTokens(): Promise<AuthTokens | null> {
    return (await this.loadSession())?.tokens ?? null;
  },

  async clearTokens() {
    await Keychain.resetGenericPassword({service: TOKEN_SERVICE});
  },

  async clearTransientData() {
    await AsyncStorage.multiRemove([ACTIVE_WORKOUT_KEY, OFFLINE_QUEUE_KEY, AI_CHAT_KEY, ...LEGACY_TRANSIENT_KEYS]);
  },

  async saveActiveWorkout(workout: WorkoutView | null) {
    await saveOwned(ACTIVE_WORKOUT_KEY, workout);
  },

  async loadActiveWorkout(): Promise<WorkoutView | null> {
    return loadOwned<WorkoutView>(ACTIVE_WORKOUT_KEY);
  },

  async saveAIChat(messages: AIChatMessage[]) {
    const recent = messages.slice(-30);
    await saveOwned(AI_CHAT_KEY, recent.length > 0 ? recent : null);
  },

  async loadAIChat(): Promise<AIChatMessage[]> {
    return (await loadOwned<AIChatMessage[]>(AI_CHAT_KEY)) ?? [];
  },

  async clearAIChat() {
    await AsyncStorage.removeItem(AI_CHAT_KEY);
  },

  async loadQueue(): Promise<OfflineOperation[]> {
    return (await loadOwned<OfflineOperation[]>(OFFLINE_QUEUE_KEY)) ?? [];
  },

  async saveQueue(items: OfflineOperation[]) {
    await saveOwned(OFFLINE_QUEUE_KEY, items.length > 0 ? items : null);
  },

  async enqueueSet(workoutId: string, payload: Extract<OfflineOperation, {type: 'log_set'}>['payload']) {
    const queue = await this.loadQueue();
    const existingIndex = queue.findIndex(
      item =>
        item.type === 'log_set' &&
        item.workoutId === workoutId &&
        item.payload.workout_exercise_id === payload.workout_exercise_id &&
        item.payload.set_number === payload.set_number,
    );
    const operation: OfflineOperation = {
      id: newLocalId(),
      type: 'log_set',
      workoutId,
      payload,
      createdAt: new Date().toISOString(),
    };
    if (existingIndex >= 0) queue[existingIndex] = operation;
    else queue.push(operation);
    await this.saveQueue(queue);
  },

  async enqueueFinish(workoutId: string) {
    const queue = await this.loadQueue();
    if (!queue.some(item => item.type === 'finish_workout' && item.workoutId === workoutId)) {
      queue.push({id: newLocalId(), type: 'finish_workout', workoutId, createdAt: new Date().toISOString()});
    }
    await this.saveQueue(queue);
  },

  async enqueueCancel(workoutId: string) {
    const queue = (await this.loadQueue()).filter(item => item.workoutId !== workoutId || item.type === 'log_set');
    if (!queue.some(item => item.type === 'cancel_workout' && item.workoutId === workoutId)) {
      queue.push({id: newLocalId(), type: 'cancel_workout', workoutId, createdAt: new Date().toISOString()});
    }
    await this.saveQueue(queue);
  },

  async removeWorkoutOperations(workoutId: string) {
    const queue = await this.loadQueue();
    await this.saveQueue(queue.filter(item => item.workoutId !== workoutId));
  },
};

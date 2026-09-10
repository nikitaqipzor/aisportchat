import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Keychain from 'react-native-keychain';
import type {AIChatMessage, AuthTokens, WorkoutView} from '../api/client';

const TOKEN_SERVICE = 'ai-fitness-os.auth';
const LEGACY_USERNAME = 'fitness-user';
const LEGACY_KEYS = ['fitness.active-workout.v1', 'fitness.offline-queue.v1', 'fitness.ai-chat.v1', 'fitness.active-workout.v2', 'fitness.offline-queue.v2', 'fitness.ai-chat.v2'];
const key = (kind: string, owner: string) => `fitness.${kind}.v3.${encodeURIComponent(owner)}`;
const revokedKey = (owner: string) => key('session-revoked', owner);
let keychainMutation: Promise<void> = Promise.resolve();

export type OfflineOperation =
  | {id: string; type: 'log_set'; workoutId: string; payload: {workout_exercise_id: string; set_number: number; weight?: number; repetitions: number; rpe?: number; rir?: number}; createdAt: string}
  | {id: string; type: 'finish_workout' | 'cancel_workout'; workoutId: string; createdAt: string};
export type StoredSession = {tokens: AuthTokens; userId?: string};
export type TrainingEnvironment = 'home' | 'gym' | 'band';
export type TrainingEnvironmentCache = {environments: TrainingEnvironment[]; cachedAt: string};

function mutateKeychain(action: () => Promise<unknown>) { const next = keychainMutation.then(action, action).then(() => undefined); keychainMutation = next.catch(() => undefined); return next; }
async function credentials() { await keychainMutation; return Keychain.getGenericPassword({service: TOKEN_SERVICE}); }
function newLocalId() { return `${Date.now()}-${Math.random().toString(36).slice(2)}`; }
export async function currentOwnerUserId(): Promise<string | undefined> { const value = await credentials(); return value && value.username !== LEGACY_USERNAME ? value.username : undefined; }

async function migrateLegacy<T>(kind: string, owner: string): Promise<T | null> {
  const oldKey = `fitness.${kind}.v2`; const raw = await AsyncStorage.getItem(oldKey); if (!raw) return null;
  try { const parsed = JSON.parse(raw) as {ownerUserId?: string; value?: T}; if (parsed.ownerUserId === owner && parsed.value !== undefined) { await AsyncStorage.setItem(key(kind, owner), JSON.stringify(parsed.value)); await AsyncStorage.removeItem(oldKey); return parsed.value; } } catch { /* discard below */ }
  await AsyncStorage.removeItem(oldKey); return null;
}
async function saveOwned<T>(kind: string, value: T | null) { const owner = await currentOwnerUserId(); if (!owner) return; const storageKey = key(kind, owner); if (value === null) await AsyncStorage.removeItem(storageKey); else await AsyncStorage.setItem(storageKey, JSON.stringify(value)); }
async function loadOwned<T>(kind: string): Promise<T | null> { const owner = await currentOwnerUserId(); if (!owner) return null; const storageKey = key(kind, owner); const raw = await AsyncStorage.getItem(storageKey); if (!raw) return migrateLegacy<T>(kind, owner); try { return JSON.parse(raw) as T; } catch { await AsyncStorage.removeItem(storageKey); return null; } }

export const sessionStorage = {
  currentUserId: currentOwnerUserId,
  async saveTokens(tokens: AuthTokens, userId?: string) { const existing = await credentials(); const current = existing && existing.username !== LEGACY_USERNAME ? existing.username : undefined; const owner = userId ?? current ?? LEGACY_USERNAME; await mutateKeychain(() => Keychain.setGenericPassword(owner, JSON.stringify(tokens), {service:TOKEN_SERVICE,accessible:Keychain.ACCESSIBLE.WHEN_UNLOCKED_THIS_DEVICE_ONLY})); if (owner !== LEGACY_USERNAME) await AsyncStorage.removeItem(revokedKey(owner)); },
  async loadSession(): Promise<StoredSession | null> { const value = await credentials(); if (!value) return null; try { return {tokens:JSON.parse(value.password) as AuthTokens,userId:value.username === LEGACY_USERNAME ? undefined : value.username}; } catch { await mutateKeychain(() => Keychain.resetGenericPassword({service:TOKEN_SERVICE})); return null; } },
  async loadTokens() { return (await this.loadSession())?.tokens ?? null; },
  async clearTokens() { await mutateKeychain(() => Keychain.resetGenericPassword({service:TOKEN_SERVICE})); },
  async revokeCurrentSession() { const owner = await currentOwnerUserId(); if (owner) await AsyncStorage.setItem(revokedKey(owner), new Date().toISOString()); await this.clearTokens(); },
  async isSessionRevoked(owner: string) { return (await AsyncStorage.getItem(revokedKey(owner))) !== null; },
  async clearTransientData(owner?: string) { const target = owner ?? await currentOwnerUserId(); if (!target) return; await Promise.all(['active-workout','offline-queue','ai-chat'].map(kind => AsyncStorage.removeItem(key(kind,target)))); await Promise.all(LEGACY_KEYS.map(item => AsyncStorage.removeItem(item))); },
  async saveTrainingEnvironments(environments: TrainingEnvironment[], initiatingOwnerUserId?: string) { const owner=initiatingOwnerUserId??await currentOwnerUserId(); const valid=[...new Set(environments.filter(value=>value==='home'||value==='gym'||value==='band'))]; if(!owner||!valid.length||await currentOwnerUserId()!==owner)return false; await AsyncStorage.setItem(key('training-environments',owner),JSON.stringify({environments:valid,cachedAt:new Date().toISOString()} satisfies TrainingEnvironmentCache)); return await currentOwnerUserId()===owner; },
  async loadTrainingEnvironments(): Promise<TrainingEnvironmentCache|null> { return loadOwned<TrainingEnvironmentCache>('training-environments'); },
  async clearTrainingEnvironments(owner?: string) { const target=owner??await currentOwnerUserId(); if(target)await AsyncStorage.removeItem(key('training-environments',target)); },
  async saveActiveWorkout(value: WorkoutView | null) { await saveOwned('active-workout', value); },
  async loadActiveWorkout() { return loadOwned<WorkoutView>('active-workout'); },
  async saveAIChat(messages: AIChatMessage[]) { const recent = messages.slice(-30); await saveOwned('ai-chat', recent.length ? recent : null); },
  async loadAIChat() { return (await loadOwned<AIChatMessage[]>('ai-chat')) ?? []; },
  async clearAIChat() { await saveOwned('ai-chat', null); },
  async loadQueue() { return (await loadOwned<OfflineOperation[]>('offline-queue')) ?? []; },
  async saveQueue(items: OfflineOperation[]) { await saveOwned('offline-queue', items.length ? items : null); },
  async enqueueSet(workoutId: string, payload: Extract<OfflineOperation,{type:'log_set'}>['payload']) { const queue=await this.loadQueue(); const index=queue.findIndex(item=>item.type==='log_set'&&item.workoutId===workoutId&&item.payload.workout_exercise_id===payload.workout_exercise_id&&item.payload.set_number===payload.set_number); const operation:OfflineOperation={id:newLocalId(),type:'log_set',workoutId,payload,createdAt:new Date().toISOString()}; if(index>=0)queue[index]=operation;else queue.push(operation);await this.saveQueue(queue); },
  async enqueueFinish(workoutId:string){const queue=await this.loadQueue();if(!queue.some(item=>item.type==='finish_workout'&&item.workoutId===workoutId))queue.push({id:newLocalId(),type:'finish_workout',workoutId,createdAt:new Date().toISOString()});await this.saveQueue(queue);},
  async enqueueCancel(workoutId:string){const queue=(await this.loadQueue()).filter(item=>item.workoutId!==workoutId||item.type==='log_set');if(!queue.some(item=>item.type==='cancel_workout'&&item.workoutId===workoutId))queue.push({id:newLocalId(),type:'cancel_workout',workoutId,createdAt:new Date().toISOString()});await this.saveQueue(queue);},
  async removeWorkoutOperations(workoutId:string){await this.saveQueue((await this.loadQueue()).filter(item=>item.workoutId!==workoutId));},
};

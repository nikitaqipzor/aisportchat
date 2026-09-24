import AsyncStorage from '@react-native-async-storage/async-storage';
import {currentOwnerUserId} from '../storage/session';

type Pending = {signature: string; key: string; createdAt: number};
const inFlight = new Map<string, {signature: string; promise: Promise<unknown>}>();

/** Keep a retry key through network timeouts, navigation and app restarts.
 * An explicit successful action clears it, so a second meal is a new action. */
export async function withFoodIdempotency<T>(kind: string, signature: string, write: (key: string) => Promise<T>): Promise<T> {
  const owner = await currentOwnerUserId();
  if (!owner) throw new Error('Для сохранения еды нужно войти в аккаунт заново.');
  const storageKey = `fitness.food-write.v1.${encodeURIComponent(owner)}.${encodeURIComponent(kind)}`;
  const active = inFlight.get(storageKey);
  if (active) {
    if (active.signature === signature) return active.promise as Promise<T>;
    try { await active.promise; } catch { /* Changing the draft starts a separate action. */ }
    return withFoodIdempotency(kind, signature, write);
  }
  const promise = (async () => {
  const raw = await AsyncStorage.getItem(storageKey);
  let pending: Pending | null = null;
  try { if (raw) pending = JSON.parse(raw) as Pending; } catch { /* Create a fresh attempt. */ }
  if (!pending || pending.signature !== signature) {
    pending = {signature, key: `food-${Date.now()}-${Math.random().toString(36).slice(2)}-${Math.random().toString(36).slice(2)}`, createdAt: Date.now()};
    await AsyncStorage.setItem(storageKey, JSON.stringify(pending));
  }
  const result = await write(pending.key);
  await AsyncStorage.removeItem(storageKey);
  return result;
  })();
  inFlight.set(storageKey, {signature, promise});
  try { return await promise; }
  finally { if (inFlight.get(storageKey)?.promise === promise) inFlight.delete(storageKey); }
}

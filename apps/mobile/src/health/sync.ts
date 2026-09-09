import {api, HealthDailySnapshot} from '../api/client';
import {currentLocalDate, localDateOffset} from '../domain/date';
import {healthConnect, MI_FITNESS_PACKAGE} from '../native/healthConnect';
import {sessionStorage} from '../storage/session';

export type HealthSyncResult = {
  processed: number;
  imported: number;
  empty: number;
  latest?: HealthDailySnapshot;
};

export async function syncHealthDays(accessToken: string, days: number, onProgress?: (processed: number, total: number) => void): Promise<HealthSyncResult> {
  const total = Math.max(1, Math.min(28, Math.round(days)));
  let processed = 0; let imported = 0; let empty = 0; let latest: HealthDailySnapshot | undefined;
  for (let offset = -(total - 1); offset <= 0; offset++) {
    const date = localDateOffset(offset);
    const native = await healthConnect.readDay(date, MI_FITNESS_PACKAGE);
    processed++; onProgress?.(processed, total);
    if (!native.data_types.length) { empty++; continue; }
    const saved = await api.importHealthSnapshot(accessToken, native as Omit<HealthDailySnapshot, 'id' | 'imported_at'>);
    latest = saved; imported++;
  }
  return {processed, imported, empty, latest};
}

// WorkManager keeps encrypted snapshots on-device. We only acknowledge a snapshot
// after the authenticated API has accepted it, so background reads are not lost
// when the phone was offline or the access token was expired.
export async function flushPassiveHealthSnapshots(accessToken: string): Promise<HealthSyncResult> {
  const ownerUserId = await sessionStorage.currentUserId();
  if (!ownerUserId) return {processed: 0, imported: 0, empty: 0};
  const pending = await healthConnect.pendingSnapshots(ownerUserId);
  let processed = 0; let imported = 0; let empty = 0; let latest: HealthDailySnapshot | undefined;
  for (const native of pending) {
    processed++;
    if (!native.data_types.length) {
      empty++;
      await healthConnect.ackPendingSnapshot(ownerUserId, native.date);
      continue;
    }
    const saved = await api.importHealthSnapshot(accessToken, native as Omit<HealthDailySnapshot, 'id' | 'imported_at'>);
    await healthConnect.ackPendingSnapshot(ownerUserId, native.date);
    latest = saved; imported++;
  }
  return {processed, imported, empty, latest};
}

export async function refreshTodayIfPossible(accessToken: string) {
  try {
    const status = await healthConnect.status();
    if (!status.permissions_granted) return {passive: 0, foreground: false};
    const passive = await flushPassiveHealthSnapshots(accessToken);
    // If WorkManager has not produced a current snapshot yet, foreground refresh
    // keeps the product fresh without waiting for the next periodic window.
    const today = await api.healthToday(accessToken, currentLocalDate()).catch(() => undefined);
    const currentFresh = today && Date.now() - new Date(today.imported_at).getTime() < 60 * 60 * 1000;
    if (!currentFresh) await syncHealthDays(accessToken, 1);
    return {passive: passive.imported, foreground: !currentFresh};
  } catch {
    return {passive: 0, foreground: false};
  }
}

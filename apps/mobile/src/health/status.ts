import type {HealthDailySnapshot} from '../api/client';
import type {HealthConnectStatus} from '../native/healthConnect';

type Snapshot = Pick<HealthDailySnapshot, 'date' | 'imported_at'>;

export function formatHealthDate(date: string): string {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(date);
  return match ? `${match[3]}.${match[2]}.${match[1]}` : date;
}

// The endpoint for a day may be empty even though an older record was imported.
// Keep that record visible, but always render it under its own date.
export function selectHealthSnapshot<T extends Snapshot>(day: string, ...candidates: Array<T | null | undefined>): T | null {
  return candidates
    .filter((candidate): candidate is T => !!candidate && /^\d{4}-\d{2}-\d{2}$/.test(candidate.date) && candidate.date <= day)
    .sort((a, b) => b.date.localeCompare(a.date) || b.imported_at.localeCompare(a.imported_at))[0] ?? null;
}

export function healthConnectionState(status: HealthConnectStatus | null, snapshot: Snapshot | null, today: string) {
  const available = status?.sdk_status === 'available';
  const granted = available && status?.permissions_granted === true;
  const isToday = snapshot?.date === today;
  const step = !status?.mi_fitness_installed ? 1 : !available ? 2 : !granted ? 3 : !snapshot ? 4 : 5;
  return {
    available,
    granted,
    isToday,
    step,
    title: step === 5 ? (isToday ? 'Данные за сегодня получены' : `Последние данные: ${formatHealthDate(snapshot!.date)}`) : `Шаг ${step} из 4`,
    badge: step === 5 ? (isToday ? 'ГОТОВО' : 'ОБНОВИТЬ') : 'НАСТРОЙКА',
    syncDetail: snapshot ? `Последняя запись: ${formatHealthDate(snapshot.date)}` : 'Синхронизируйте данные за сегодня',
  };
}

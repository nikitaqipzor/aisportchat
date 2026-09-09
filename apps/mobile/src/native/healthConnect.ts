import NativeHealthConnect from '../../specs/NativeHealthConnect';

export const MI_FITNESS_PACKAGE = 'com.xiaomi.wearable';

export type PassiveSyncStatus = {
  enabled: boolean;
  last_success_at?: string | null;
  last_error?: string;
  pending_snapshots: number;
};

export type HealthConnectStatus = {
  sdk_status: 'available' | 'update_required' | 'unavailable';
  permissions_granted: boolean;
  granted_permissions: string[];
  missing_permissions: string[];
  mi_fitness_installed: boolean;
  background_read_available: boolean;
  background_read_granted: boolean;
  passive_sync: PassiveSyncStatus;
};

export type NativeHealthSnapshot = {
  date: string;
  provider: 'health_connect';
  source_package: string;
  source_label: string;
  steps: number;
  distance_m: number;
  active_calories_kcal: number;
  sleep_minutes: number;
  deep_sleep_minutes: number;
  light_sleep_minutes: number;
  rem_sleep_minutes: number;
  awake_minutes: number;
  exercise_minutes: number;
  exercise_sessions: number;
  exercise_heart_rate_avg?: number;
  exercise_heart_rate_max?: number;
  data_types: string[];
  captured_at: string;
};

const unavailableStatus: HealthConnectStatus = {
  sdk_status: 'unavailable', permissions_granted: false, granted_permissions: [], missing_permissions: [], mi_fitness_installed: false,
  background_read_available: false, background_read_granted: false,
  passive_sync: {enabled: false, pending_snapshots: 0},
};

export const healthConnect = {
  available: () => NativeHealthConnect != null,
  async status(): Promise<HealthConnectStatus> {
    if (!NativeHealthConnect) return unavailableStatus;
    return JSON.parse(await NativeHealthConnect.getStatus()) as HealthConnectStatus;
  },
  async requestPermissions(): Promise<HealthConnectStatus> {
    if (!NativeHealthConnect) throw new Error('Health Connect доступен только в Android native build');
    return JSON.parse(await NativeHealthConnect.requestPermissions()) as HealthConnectStatus;
  },
  async requestBackgroundPermission(): Promise<HealthConnectStatus> {
    if (!NativeHealthConnect) throw new Error('Фоновое чтение Health Connect доступно только в Android native build');
    return JSON.parse(await NativeHealthConnect.requestBackgroundPermission()) as HealthConnectStatus;
  },
  async readDay(date: string, sourcePackage = MI_FITNESS_PACKAGE): Promise<NativeHealthSnapshot> {
    if (!NativeHealthConnect) throw new Error('Health Connect доступен только в Android native build');
    return JSON.parse(await NativeHealthConnect.readDailySnapshot(date, sourcePackage)) as NativeHealthSnapshot;
  },
  async setPassiveSyncEnabled(ownerUserId: string, enabled: boolean): Promise<HealthConnectStatus> {
    if (!NativeHealthConnect) throw new Error('Фоновая синхронизация доступна только в Android native build');
    if (!ownerUserId.trim()) throw new Error('Для фоновой синхронизации требуется авторизованный пользователь');
    return JSON.parse(await NativeHealthConnect.setPassiveSyncEnabled(ownerUserId, enabled)) as HealthConnectStatus;
  },
  async pendingSnapshots(ownerUserId: string): Promise<NativeHealthSnapshot[]> {
    if (!NativeHealthConnect || !ownerUserId.trim()) return [];
    return JSON.parse(await NativeHealthConnect.getPendingSnapshots(ownerUserId)) as NativeHealthSnapshot[];
  },
  async ackPendingSnapshot(ownerUserId: string, date: string) {
    if (!NativeHealthConnect || !ownerUserId.trim()) return false;
    return NativeHealthConnect.ackPendingSnapshot(ownerUserId, date);
  },
  async clearPassiveSync(ownerUserId: string) {
    if (!NativeHealthConnect || !ownerUserId.trim()) return false;
    return NativeHealthConnect.clearPassiveSync(ownerUserId);
  },
  async bindSessionOwner(ownerUserId: string) {
    if (!NativeHealthConnect || !ownerUserId.trim()) return false;
    return NativeHealthConnect.bindSessionOwner(ownerUserId);
  },
  async openSettings() {
    if (!NativeHealthConnect) return false;
    return NativeHealthConnect.openSettings();
  },
};

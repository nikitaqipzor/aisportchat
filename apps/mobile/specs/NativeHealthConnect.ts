import type {TurboModule} from 'react-native';
import {TurboModuleRegistry} from 'react-native';

export interface Spec extends TurboModule {
  getStatus(): Promise<string>;
  requestPermissions(): Promise<string>;
  requestBackgroundPermission(): Promise<string>;
  readDailySnapshot(date: string, preferredSourcePackage: string): Promise<string>;
  openSettings(): Promise<boolean>;
  setPassiveSyncEnabled(ownerUserId: string, enabled: boolean): Promise<string>;
  getPendingSnapshots(ownerUserId: string): Promise<string>;
  ackPendingSnapshot(ownerUserId: string, date: string): Promise<boolean>;
  clearPassiveSync(ownerUserId: string): Promise<boolean>;
  bindSessionOwner(ownerUserId: string): Promise<boolean>;
}

export default TurboModuleRegistry.get<Spec>('HealthConnect');

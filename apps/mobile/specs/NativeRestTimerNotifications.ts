import type {TurboModule} from 'react-native';
import {TurboModuleRegistry} from 'react-native';

export interface Spec extends TurboModule {
  schedule(seconds: number, exerciseName: string): Promise<boolean>;
  cancel(): Promise<boolean>;
}

export default TurboModuleRegistry.get<Spec>('RestTimerNotifications');

import type {TurboModule} from 'react-native';
import {TurboModuleRegistry} from 'react-native';

export interface Spec extends TurboModule {
  startLiveSession(exerciseKey: string, maxDurationSeconds: number): Promise<string>;
}

export default TurboModuleRegistry.get<Spec>('TechniqueLive');

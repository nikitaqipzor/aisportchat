import type {TurboModule} from 'react-native';
import {TurboModuleRegistry} from 'react-native';

export interface Spec extends TurboModule {
  getApiBaseUrl(): string;
  clearPrivateTempFiles(): Promise<boolean>;
}

export default TurboModuleRegistry.get<Spec>('AIFitnessConfig');

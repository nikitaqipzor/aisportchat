import type {TurboModule} from 'react-native';
import {TurboModuleRegistry} from 'react-native';

export interface Spec extends TurboModule {
  recordVideo(maxDurationSeconds: number): Promise<string>;
  analyzeVideo(filePath: string, intervalMs: number): Promise<string>;
  deleteVideo(filePath: string): Promise<boolean>;
}

export default TurboModuleRegistry.get<Spec>('TechniqueVideo');

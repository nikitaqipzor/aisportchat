import {Platform} from 'react-native';
import NativeAIFitnessConfig from '../../specs/NativeAIFitnessConfig';

export async function clearPrivateTempFiles() {
  if (Platform.OS !== 'android' || !NativeAIFitnessConfig) return true;
  return NativeAIFitnessConfig.clearPrivateTempFiles();
}

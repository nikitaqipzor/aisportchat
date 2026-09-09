import {NativeModules} from 'react-native';
import NativeAIFitnessConfig from '../../specs/NativeAIFitnessConfig';

declare const __DEV__: boolean;
declare global {
  // Can be assigned before the app bundle starts by a native bootstrap/config layer.
  // eslint-disable-next-line no-var
  var __AI_FITNESS_API_BASE_URL__: string | undefined;
}

function normalize(value: string) {
  return value.replace(/\/+$/, '');
}

function validateConfiguredUrl(value: string) {
  const normalized = normalize(value);
  if (!__DEV__ && !normalized.startsWith('https://')) {
    throw new Error('Production API URL must use HTTPS.');
  }
  return normalized;
}

function metroHost(): string | undefined {
  const sourceCode = NativeModules.SourceCode as {scriptURL?: string} | undefined;
  const scriptURL = sourceCode?.scriptURL;
  if (!scriptURL) return undefined;
  const match = scriptURL.match(/^https?:\/\/([^/:]+)(?::\d+)?\//i);
  return match?.[1];
}

export function resolveApiBaseUrl() {
  const runtime = globalThis.__AI_FITNESS_API_BASE_URL__;
  if (runtime) return validateConfiguredUrl(runtime);

  const nativeUrl = NativeAIFitnessConfig?.getApiBaseUrl();
  if (nativeUrl) return validateConfiguredUrl(nativeUrl);

  if (__DEV__) {
    const host = metroHost() ?? '10.0.2.2';
    return `http://${host}:8080/api/v1`;
  }

  throw new Error('Production API URL is not configured. Set AI_FITNESS_API_BASE_URL for the release build.');
}

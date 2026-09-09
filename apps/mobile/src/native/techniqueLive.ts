import NativeTechniqueLive from '../../specs/NativeTechniqueLive';
import type {NativePoseVideo} from './techniqueVideo';

export type NativeLivePoseVideo = NativePoseVideo & {
  capture_mode: 'live';
  exercise_key: string;
  live_rep_count: number;
};

export const techniqueLive = {
  available: () => NativeTechniqueLive != null,
  async start(exerciseKey: string, maxDurationSeconds = 45): Promise<NativeLivePoseVideo> {
    if (!NativeTechniqueLive) throw new Error('Live Technique доступен только в Android native build');
    return JSON.parse(await NativeTechniqueLive.startLiveSession(exerciseKey, maxDurationSeconds)) as NativeLivePoseVideo;
  },
};

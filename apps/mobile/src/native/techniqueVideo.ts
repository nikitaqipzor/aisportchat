import NativeTechniqueVideo from '../../specs/NativeTechniqueVideo';

export type PoseLandmark = {x: number; y: number; z: number; visibility: number};
export type PoseFrame = {timestamp_ms: number; landmarks: PoseLandmark[]};
export type NativePoseVideo = {duration_ms: number; frame_interval_ms: number; width: number; height: number; valid_frames: number; frames: PoseFrame[]; model: string};

export const techniqueVideo = {
  available: () => NativeTechniqueVideo != null,
  async record(maxDurationSeconds = 45) {
    if (!NativeTechniqueVideo) throw new Error('Technique Video доступен только в Android native build');
    return NativeTechniqueVideo.recordVideo(maxDurationSeconds);
  },
  async analyze(path: string, intervalMs = 120): Promise<NativePoseVideo> {
    if (!NativeTechniqueVideo) throw new Error('Technique Video доступен только в Android native build');
    return JSON.parse(await NativeTechniqueVideo.analyzeVideo(path, intervalMs)) as NativePoseVideo;
  },
  async cleanup(path: string) { if (NativeTechniqueVideo) await NativeTechniqueVideo.deleteVideo(path); },
};

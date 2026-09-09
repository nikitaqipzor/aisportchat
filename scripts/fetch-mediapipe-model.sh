#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/apps/mobile/android/app/src/main/assets/pose_landmarker_lite.task"
URL="https://storage.googleapis.com/mediapipe-models/pose_landmarker/pose_landmarker_lite/float16/1/pose_landmarker_lite.task"
mkdir -p "$(dirname "$OUT")"
if [ -s "$OUT" ]; then echo "Pose model already exists: $OUT"; exit 0; fi
if command -v curl >/dev/null 2>&1; then curl -fL --retry 3 "$URL" -o "$OUT"; elif command -v wget >/dev/null 2>&1; then wget -O "$OUT" "$URL"; else echo "curl or wget is required" >&2; exit 1; fi
[ -s "$OUT" ] || { echo "model download failed" >&2; exit 1; }
echo "Downloaded: $OUT"

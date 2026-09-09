package com.aifitnessos.technique

import android.app.Activity
import android.content.Intent
import android.media.MediaMetadataRetriever
import android.provider.MediaStore
import androidx.core.content.FileProvider
import com.aifitnessos.specs.NativeTechniqueVideoSpec
import com.facebook.react.bridge.ActivityEventListener
import com.facebook.react.bridge.BaseActivityEventListener
import com.facebook.react.bridge.Promise
import com.facebook.react.bridge.ReactApplicationContext
import com.google.mediapipe.framework.image.BitmapImageBuilder
import com.google.mediapipe.tasks.core.BaseOptions
import com.google.mediapipe.tasks.vision.core.RunningMode
import com.google.mediapipe.tasks.vision.poselandmarker.PoseLandmarker
import org.json.JSONArray
import org.json.JSONObject
import java.io.File
import java.util.concurrent.Executors

class TechniqueVideoModule(reactContext: ReactApplicationContext) : NativeTechniqueVideoSpec(reactContext) {
    private var pendingPromise: Promise? = null
    private var pendingFile: File? = null
    private val executor = Executors.newSingleThreadExecutor()

    private val listener: ActivityEventListener = object : BaseActivityEventListener() {
        override fun onActivityResult(activity: Activity, requestCode: Int, resultCode: Int, data: Intent?) {
            if (requestCode != REQUEST_VIDEO) return
            val promise = pendingPromise ?: return
            val file = pendingFile
            pendingPromise = null
            pendingFile = null
            if (resultCode != Activity.RESULT_OK || file == null || !file.exists() || file.length() == 0L) {
                file?.delete()
                promise.reject("TECHNIQUE_VIDEO_CANCELLED", "Запись видео отменена")
                return
            }
            promise.resolve(file.absolutePath)
        }
    }

    init { reactApplicationContext.addActivityEventListener(listener) }
    override fun getName() = NAME

    override fun recordVideo(maxDurationSeconds: Double, promise: Promise) {
        if (pendingPromise != null) { promise.reject("TECHNIQUE_VIDEO_BUSY", "Камера уже открыта"); return }
        val activity = reactApplicationContext.currentActivity ?: run { promise.reject("TECHNIQUE_VIDEO_NO_ACTIVITY", "Нет активного Android Activity"); return }
        val intent = Intent(MediaStore.ACTION_VIDEO_CAPTURE)
        if (intent.resolveActivity(reactApplicationContext.packageManager) == null) { promise.reject("TECHNIQUE_VIDEO_NO_CAMERA", "Видео-камера недоступна"); return }
        val dir = File(reactApplicationContext.cacheDir, "technique_video").apply { mkdirs() }
        val file = File.createTempFile("technique_", ".mp4", dir)
        val uri = FileProvider.getUriForFile(reactApplicationContext, "${reactApplicationContext.packageName}.fileprovider", file)
        intent.putExtra(MediaStore.EXTRA_OUTPUT, uri)
        intent.putExtra(MediaStore.EXTRA_VIDEO_QUALITY, 1)
        intent.putExtra(MediaStore.EXTRA_DURATION_LIMIT, maxDurationSeconds.toInt().coerceIn(5, 90))
        intent.addFlags(Intent.FLAG_GRANT_WRITE_URI_PERMISSION or Intent.FLAG_GRANT_READ_URI_PERMISSION)
        pendingPromise = promise
        pendingFile = file
        activity.startActivityForResult(intent, REQUEST_VIDEO)
    }

    override fun analyzeVideo(filePath: String, intervalMs: Double, promise: Promise) {
        val file = File(filePath)
        if (!file.exists()) { promise.reject("TECHNIQUE_VIDEO_MISSING", "Видео не найдено"); return }
        val interval = intervalMs.toLong().coerceIn(80L, 500L)
        executor.execute {
            try { promise.resolve(analyze(file, interval).toString()) }
            catch (error: Throwable) { promise.reject("TECHNIQUE_ANALYSIS_FAILED", error) }
        }
    }

    override fun deleteVideo(filePath: String, promise: Promise) {
        val f = File(filePath)
        promise.resolve(!f.exists() || f.delete())
    }

    private fun analyze(file: File, intervalMs: Long): JSONObject {
        ensureModelExists()
        val retriever = MediaMetadataRetriever()
        retriever.setDataSource(file.absolutePath)
        val durationMs = retriever.extractMetadata(MediaMetadataRetriever.METADATA_KEY_DURATION)?.toLongOrNull()
            ?: error("Не удалось определить длительность видео")
        if (durationMs <= 0 || durationMs > MAX_DURATION_MS) error("Видео должно быть короче 90 секунд")

        val options = PoseLandmarker.PoseLandmarkerOptions.builder()
            .setBaseOptions(BaseOptions.builder().setModelAssetPath(MODEL_ASSET).build())
            .setRunningMode(RunningMode.VIDEO)
            .setMinPoseDetectionConfidence(0.5f)
            .setMinPosePresenceConfidence(0.5f)
            .setMinTrackingConfidence(0.5f)
            .build()
        val landmarker = PoseLandmarker.createFromOptions(reactApplicationContext, options)
        val framesJSON = JSONArray()
        var validFrames = 0
        var timestamp = 0L
        var width = 0
        var height = 0
        try {
            while (timestamp <= durationMs) {
                val frame = retriever.getFrameAtTime(timestamp * 1000, MediaMetadataRetriever.OPTION_CLOSEST)
                if (frame != null) {
                    if (width == 0) { width = frame.width; height = frame.height }
                    val argb = if (frame.config == android.graphics.Bitmap.Config.ARGB_8888) frame else frame.copy(android.graphics.Bitmap.Config.ARGB_8888, false)
                    val image = BitmapImageBuilder(argb).build()
                    val result = landmarker.detectForVideo(image, timestamp)
                    val pose = result.landmarks().firstOrNull()
                    if (pose != null && pose.size >= 33) {
                        val landmarks = JSONArray()
                        for (point in pose) {
                            landmarks.put(JSONObject().apply {
                                put("x", point.x().toDouble()); put("y", point.y().toDouble()); put("z", point.z().toDouble())
                                put("visibility", if (point.visibility().isPresent) point.visibility().get().toDouble() else 0.0)
                            })
                        }
                        framesJSON.put(JSONObject().put("timestamp_ms", timestamp).put("landmarks", landmarks)); validFrames++
                    }
                    if (argb !== frame) argb.recycle()
                    frame.recycle()
                }
                timestamp += intervalMs
            }
        } finally { landmarker.close(); retriever.release() }
        if (validFrames < 8) error("Не удалось устойчиво распознать позу. Убедитесь, что всё тело видно в кадре")
        return JSONObject().apply {
            put("duration_ms", durationMs); put("frame_interval_ms", intervalMs); put("width", width); put("height", height)
            put("valid_frames", validFrames); put("frames", framesJSON); put("model", MODEL_ASSET)
        }
    }

    private fun ensureModelExists() {
        try { reactApplicationContext.assets.open(MODEL_ASSET).close() }
        catch (_: Throwable) { error("Pose Landmarker model отсутствует. Запустите scripts/fetch-mediapipe-model.sh перед Android build") }
    }

    companion object {
        const val NAME = "TechniqueVideo"
        private const val REQUEST_VIDEO = 4321
        private const val MODEL_ASSET = "pose_landmarker_lite.task"
        private const val MAX_DURATION_MS = 90_000L
    }
}

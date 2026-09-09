package com.aifitnessos.technique

import android.Manifest
import android.app.Activity
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.Color
import android.graphics.Matrix
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.os.SystemClock
import android.view.Gravity
import android.view.View
import android.view.WindowManager
import android.widget.Button
import android.widget.FrameLayout
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.camera.core.CameraSelector
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.ImageProxy
import androidx.camera.core.Preview
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.core.content.ContextCompat
import com.google.mediapipe.framework.image.BitmapImageBuilder
import com.google.mediapipe.tasks.core.BaseOptions
import com.google.mediapipe.tasks.vision.core.RunningMode
import com.google.mediapipe.tasks.vision.poselandmarker.PoseLandmarker
import org.json.JSONArray
import org.json.JSONObject
import java.io.File
import java.util.concurrent.ExecutorService
import java.util.concurrent.Executors
import kotlin.math.max

class TechniqueLiveActivity : AppCompatActivity() {
    private lateinit var previewView: PreviewView
    private lateinit var overlayView: TechniquePoseOverlayView
    private lateinit var repText: TextView
    private lateinit var statusText: TextView
    private lateinit var timeText: TextView
    private lateinit var switchButton: Button
    private lateinit var stopButton: Button

    private var cameraProvider: ProcessCameraProvider? = null
    private lateinit var analyzerExecutor: ExecutorService
    private var poseLandmarker: PoseLandmarker? = null
    private lateinit var repCounter: LiveRepCounter
    private lateinit var exerciseKey: String
    private var useFrontCamera = false
    private var maxDurationSeconds = 45
    private var sessionStartMs = 0L
    private var lastInferenceAt = 0L
    private var lastSampleAt = -SAMPLE_INTERVAL_MS
    private var validFrames = 0
    private var lastWidth = 0
    private var lastHeight = 0
    private var finishing = false
    private val samples = mutableListOf<LivePoseSample>()
    private val handler = Handler(Looper.getMainLooper())
    private val autoFinish = Runnable { finishSession() }
    private val clockTick = object : Runnable {
        override fun run() {
            if (sessionStartMs > 0 && !finishing) {
                val elapsed = (SystemClock.elapsedRealtime() - sessionStartMs).coerceAtLeast(0)
                timeText.text = "%d:%02d".format(elapsed / 60_000, (elapsed / 1000) % 60)
                handler.postDelayed(this, 500)
            }
        }
    }

    private val permissionLauncher = registerForActivityResult(ActivityResultContracts.RequestPermission()) { granted ->
        if (granted) startCamera() else finishError("Для live-анализа нужен доступ к камере")
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
        exerciseKey = intent.getStringExtra(EXTRA_EXERCISE_KEY).orEmpty()
        maxDurationSeconds = intent.getIntExtra(EXTRA_MAX_DURATION, 45).coerceIn(10, 60)
        if (exerciseKey !in SUPPORTED) {
            finishError("Упражнение пока не поддерживается live-анализом")
            return
        }
        repCounter = LiveRepCounter(exerciseKey)
        analyzerExecutor = Executors.newSingleThreadExecutor()
        buildUI()
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED) {
            startCamera()
        } else {
            permissionLauncher.launch(Manifest.permission.CAMERA)
        }
    }

    private fun buildUI() {
        window.statusBarColor = Color.BLACK
        window.navigationBarColor = Color.BLACK
        val root = FrameLayout(this).apply { setBackgroundColor(Color.BLACK) }
        previewView = PreviewView(this).apply {
            scaleType = PreviewView.ScaleType.FIT_CENTER
            implementationMode = PreviewView.ImplementationMode.COMPATIBLE
            setBackgroundColor(Color.BLACK)
        }
        overlayView = TechniquePoseOverlayView(this)
        root.addView(previewView, FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT, FrameLayout.LayoutParams.MATCH_PARENT))
        root.addView(overlayView, FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT, FrameLayout.LayoutParams.MATCH_PARENT))

        val top = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(16), dp(12), dp(16), dp(12))
            setBackgroundColor(0xB0000000.toInt())
        }
        val title = TextView(this).apply {
            text = "LIVE · ${exerciseName(exerciseKey)}"
            setTextColor(Color.WHITE); textSize = 16f; setTypeface(typeface, android.graphics.Typeface.BOLD)
        }
        val row = LinearLayout(this).apply { orientation = LinearLayout.HORIZONTAL; gravity = Gravity.CENTER_VERTICAL }
        repText = TextView(this).apply {
            text = "0 повторений"; setTextColor(Color.WHITE); textSize = 28f; setTypeface(typeface, android.graphics.Typeface.BOLD)
        }
        timeText = TextView(this).apply { text = "0:00"; setTextColor(0xFFBDBDBD.toInt()); textSize = 18f; gravity = Gravity.END }
        row.addView(repText, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f))
        row.addView(timeText, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, .35f))
        statusText = TextView(this).apply {
            text = exerciseHint(exerciseKey); setTextColor(0xFFE0E0E0.toInt()); textSize = 13f; setPadding(0, dp(6), 0, 0)
        }
        top.addView(title); top.addView(row); top.addView(statusText)
        root.addView(top, FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT, FrameLayout.LayoutParams.WRAP_CONTENT, Gravity.TOP))

        val bottom = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER
            setPadding(dp(14), dp(10), dp(14), dp(18))
            setBackgroundColor(0xB0000000.toInt())
        }
        switchButton = Button(this).apply {
            text = "Сменить камеру"
            setOnClickListener { switchCamera() }
        }
        stopButton = Button(this).apply {
            text = "Завершить анализ"
            setOnClickListener { finishSession() }
        }
        bottom.addView(switchButton, LinearLayout.LayoutParams(0, dp(52), 1f).apply { marginEnd = dp(8) })
        bottom.addView(stopButton, LinearLayout.LayoutParams(0, dp(52), 1.3f))
        root.addView(bottom, FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT, FrameLayout.LayoutParams.WRAP_CONTENT, Gravity.BOTTOM))
        setContentView(root)
    }

    private fun startCamera() {
        try {
            ensureModel()
            if (poseLandmarker == null) {
                poseLandmarker = PoseLandmarker.createFromOptions(
                    this,
                    PoseLandmarker.PoseLandmarkerOptions.builder()
                        .setBaseOptions(BaseOptions.builder().setModelAssetPath(MODEL_ASSET).build())
                        .setRunningMode(RunningMode.IMAGE)
                        .setMinPoseDetectionConfidence(.5f)
                        .setMinPosePresenceConfidence(.5f)
                        .setMinTrackingConfidence(.5f)
                        .build(),
                )
            }
        } catch (error: Throwable) {
            finishError(error.message ?: "Не удалось запустить Pose Landmarker")
            return
        }
        val future = ProcessCameraProvider.getInstance(this)
        future.addListener({
            try {
                cameraProvider = future.get()
                bindCamera()
                if (sessionStartMs == 0L) {
                    resetAnalysisState()
                    handler.post(clockTick)
                    handler.postDelayed(autoFinish, maxDurationSeconds * 1000L)
                }
            } catch (error: Throwable) {
                finishError(error.message ?: "Не удалось открыть камеру")
            }
        }, ContextCompat.getMainExecutor(this))
    }

    private fun bindCamera() {
        val provider = cameraProvider ?: return
        provider.unbindAll()
        val selector = if (useFrontCamera) CameraSelector.DEFAULT_FRONT_CAMERA else CameraSelector.DEFAULT_BACK_CAMERA
        val preview = Preview.Builder().build().also { it.setSurfaceProvider(previewView.surfaceProvider) }
        val analysis = ImageAnalysis.Builder()
            .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
            .setOutputImageFormat(ImageAnalysis.OUTPUT_IMAGE_FORMAT_RGBA_8888)
            .build()
        analysis.setAnalyzer(analyzerExecutor) { proxy -> analyzeFrame(proxy) }
        provider.bindToLifecycle(this, selector, preview, analysis)
        switchButton.visibility = if (hasBothCameras(provider)) View.VISIBLE else View.GONE
    }

    private fun hasBothCameras(provider: ProcessCameraProvider): Boolean = try {
        provider.hasCamera(CameraSelector.DEFAULT_BACK_CAMERA) && provider.hasCamera(CameraSelector.DEFAULT_FRONT_CAMERA)
    } catch (_: Throwable) { false }

    private fun switchCamera() {
        useFrontCamera = !useFrontCamera
        resetAnalysisState()
        handler.removeCallbacks(autoFinish)
        handler.postDelayed(autoFinish, maxDurationSeconds * 1000L)
        bindCamera()
    }

    private fun resetAnalysisState() {
        synchronized(samples) { samples.clear() }
        repCounter = LiveRepCounter(exerciseKey)
        validFrames = 0
        lastSampleAt = -SAMPLE_INTERVAL_MS
        lastInferenceAt = 0L
        sessionStartMs = SystemClock.elapsedRealtime()
        repText.text = "0 повторений"
        overlayView.clearPose()
        statusText.text = exerciseHint(exerciseKey)
    }

    private fun analyzeFrame(proxy: ImageProxy) {
        try {
            val now = SystemClock.elapsedRealtime()
            if (now - lastInferenceAt < INFERENCE_INTERVAL_MS) return
            lastInferenceAt = now
            val raw = rgbaBitmap(proxy)
            val rotated = rotateBitmap(raw, proxy.imageInfo.rotationDegrees)
            val brightness = estimateBrightness(rotated)
            val image = BitmapImageBuilder(rotated).build()
            val result = poseLandmarker?.detect(image)
            val pose = result?.landmarks()?.firstOrNull()
            val timestamp = (now - sessionStartMs).coerceAtLeast(0)
            if (pose != null && pose.size >= 33) {
                validFrames++
                val points = pose.map { p ->
                    LivePosePoint(
                        p.x(), p.y(), p.z(),
                        if (p.visibility().isPresent) p.visibility().get() else 0f,
                    )
                }
                val reps = repCounter.update(points, timestamp)
                if (timestamp - lastSampleAt >= SAMPLE_INTERVAL_MS) {
                    synchronized(samples) {
                        if (samples.size < MAX_SAMPLES) samples.add(LivePoseSample(timestamp, points))
                    }
                    lastSampleAt = timestamp
                }
                lastWidth = rotated.width
                lastHeight = rotated.height
                val guidance = framingGuidance(points, brightness)
                runOnUiThread {
                    overlayView.updatePose(points, rotated.width, rotated.height, useFrontCamera)
                    repText.text = "$reps ${repWord(reps)}"
                    statusText.text = guidance
                }
            } else {
                runOnUiThread {
                    overlayView.clearPose()
                    statusText.text = if (brightness < 45) "Слишком темно — добавьте света" else "Не вижу всё тело — отойдите дальше и встаньте в кадр"
                }
            }
            if (rotated !== raw) raw.recycle()
            rotated.recycle()
        } catch (_: Throwable) {
            runOnUiThread { statusText.text = "Не удаётся устойчиво распознать позу. Проверьте свет и ракурс." }
        } finally {
            proxy.close()
        }
    }

    private fun rgbaBitmap(proxy: ImageProxy): Bitmap {
        val plane = proxy.planes[0]
        val paddedWidth = max(proxy.width, plane.rowStride / plane.pixelStride.coerceAtLeast(1))
        val padded = Bitmap.createBitmap(paddedWidth, proxy.height, Bitmap.Config.ARGB_8888)
        plane.buffer.rewind()
        padded.copyPixelsFromBuffer(plane.buffer)
        if (paddedWidth == proxy.width) return padded
        return Bitmap.createBitmap(padded, 0, 0, proxy.width, proxy.height).also { padded.recycle() }
    }

    private fun rotateBitmap(source: Bitmap, degrees: Int): Bitmap {
        if (degrees % 360 == 0) return source
        return Bitmap.createBitmap(source, 0, 0, source.width, source.height, Matrix().apply { postRotate(degrees.toFloat()) }, true)
    }

    private fun estimateBrightness(bitmap: Bitmap): Int {
        var sum = 0L
        var n = 0
        val stepX = max(1, bitmap.width / 16)
        val stepY = max(1, bitmap.height / 24)
        var y = 0
        while (y < bitmap.height) {
            var x = 0
            while (x < bitmap.width) {
                val c = bitmap.getPixel(x, y)
                sum += (Color.red(c) * 299 + Color.green(c) * 587 + Color.blue(c) * 114) / 1000
                n++
                x += stepX
            }
            y += stepY
        }
        return if (n == 0) 0 else (sum / n).toInt()
    }

    private fun framingGuidance(points: List<LivePosePoint>, brightness: Int): String {
        if (brightness < 45) return "Слишком темно — добавьте света"
        val key = intArrayOf(0, 11, 12, 15, 16, 23, 24, 25, 26, 27, 28)
        val visible = key.mapNotNull { index -> points.getOrNull(index)?.takeIf { it.visibility >= .5f } }
        if (visible.size < 8) return "Не вижу всё тело — отойдите дальше или измените ракурс"
        val confidence = visible.map { it.visibility }.average()
        val minX = visible.minOf { it.x }; val maxX = visible.maxOf { it.x }
        val minY = visible.minOf { it.y }; val maxY = visible.maxOf { it.y }
        if (minX < .035f || maxX > .965f || minY < .025f || maxY > .985f) return "Отойдите дальше — тело касается края кадра"
        if (max(maxX - minX, maxY - minY) < .52f) return "Можно подойти немного ближе — вы слишком далеко"
        if (confidence < .65) return "Поза видна неуверенно — улучшите свет и не закрывайте суставы"
        return "Кадр хороший · ${exerciseHint(exerciseKey)}"
    }

    private fun finishSession() {
        if (finishing) return
        finishing = true
        handler.removeCallbacks(autoFinish)
        handler.removeCallbacks(clockTick)
        cameraProvider?.unbindAll()
        val snapshot = synchronized(samples) { samples.toList() }
        if (snapshot.size < 8) {
            finishError("Недостаточно устойчивых кадров. Запишите минимум несколько полных повторений, удерживая всё тело в кадре.")
            return
        }
        try {
            val json = JSONObject().apply {
                put("capture_mode", "live")
                put("exercise_key", exerciseKey)
                put("duration_ms", (SystemClock.elapsedRealtime() - sessionStartMs).coerceAtLeast(0))
                put("frame_interval_ms", SAMPLE_INTERVAL_MS)
                put("width", lastWidth)
                put("height", lastHeight)
                put("valid_frames", validFrames)
                put("live_rep_count", repCounter.count)
                put("model", MODEL_ASSET)
                put("frames", JSONArray().apply {
                    snapshot.forEach { sample ->
                        put(JSONObject().apply {
                            put("timestamp_ms", sample.timestampMs)
                            put("landmarks", JSONArray().apply {
                                sample.points.forEach { p ->
                                    put(JSONObject().apply {
                                        put("x", p.x.toDouble()); put("y", p.y.toDouble()); put("z", p.z.toDouble()); put("visibility", p.visibility.toDouble())
                                    })
                                }
                            })
                        })
                    }
                })
            }
            val dir = File(cacheDir, "technique_live").apply { mkdirs() }
            val file = File.createTempFile("live_", ".json", dir).apply { writeText(json.toString()) }
            setResult(Activity.RESULT_OK, Intent().putExtra(EXTRA_RESULT_PATH, file.absolutePath))
            finish()
        } catch (error: Throwable) {
            finishError(error.message ?: "Не удалось сохранить результат live-анализа")
        }
    }

    private fun finishError(message: String) {
        if (!isFinishing) {
            setResult(Activity.RESULT_CANCELED, Intent().putExtra(EXTRA_ERROR, message))
            finish()
        }
    }

    private fun ensureModel() {
        try { assets.open(MODEL_ASSET).close() }
        catch (_: Throwable) { error("Pose Landmarker model отсутствует") }
    }

    override fun onDestroy() {
        handler.removeCallbacks(autoFinish)
        handler.removeCallbacks(clockTick)
        cameraProvider?.unbindAll()
        poseLandmarker?.close()
        if (::analyzerExecutor.isInitialized) analyzerExecutor.shutdownNow()
        super.onDestroy()
    }

    private fun dp(value: Int): Int = (value * resources.displayMetrics.density).toInt()

    companion object {
        const val EXTRA_EXERCISE_KEY = "exercise_key"
        const val EXTRA_MAX_DURATION = "max_duration_seconds"
        const val EXTRA_RESULT_PATH = "result_path"
        const val EXTRA_ERROR = "error"
        private const val MODEL_ASSET = "pose_landmarker_lite.task"
        private const val INFERENCE_INTERVAL_MS = 110L
        private const val SAMPLE_INTERVAL_MS = 200L
        private const val MAX_SAMPLES = 320
        private val SUPPORTED = setOf("squat", "biceps_curl", "push_up", "lunge", "shoulder_press")

        private fun exerciseName(key: String) = when (key) {
            "squat" -> "Приседания"
            "biceps_curl" -> "Подъём на бицепс"
            "push_up" -> "Отжимания"
            "lunge" -> "Выпады"
            "shoulder_press" -> "Жим над головой"
            else -> "Упражнение"
        }
        private fun exerciseHint(key: String) = when (key) {
            "squat", "lunge" -> "Лучший ракурс: сбоку или 45°, камера примерно на уровне таза."
            "push_up" -> "Поставьте телефон сбоку так, чтобы кисти, плечи, таз и стопы были видны."
            "biceps_curl" -> "Поставьте камеру спереди под небольшим углом, локти и кисти не должны выходить из кадра."
            "shoulder_press" -> "Камера спереди или 45°, в кадре должны оставаться кисти даже в верхней точке."
            else -> "Всё тело и работающие суставы должны оставаться в кадре."
        }
        private fun repWord(count: Int) = when {
            count % 10 == 1 && count % 100 != 11 -> "повторение"
            count % 10 in 2..4 && count % 100 !in 12..14 -> "повторения"
            else -> "повторений"
        }
    }
}

data class LivePoseSample(val timestampMs: Long, val points: List<LivePosePoint>)

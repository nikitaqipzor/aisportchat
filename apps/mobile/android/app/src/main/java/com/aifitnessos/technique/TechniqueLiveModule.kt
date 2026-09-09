package com.aifitnessos.technique

import android.app.Activity
import android.content.Intent
import com.aifitnessos.specs.NativeTechniqueLiveSpec
import com.facebook.react.bridge.ActivityEventListener
import com.facebook.react.bridge.BaseActivityEventListener
import com.facebook.react.bridge.Promise
import com.facebook.react.bridge.ReactApplicationContext
import java.io.File

class TechniqueLiveModule(reactContext: ReactApplicationContext) : NativeTechniqueLiveSpec(reactContext) {
    private var pendingPromise: Promise? = null

    private val listener: ActivityEventListener = object : BaseActivityEventListener() {
        override fun onActivityResult(activity: Activity, requestCode: Int, resultCode: Int, data: Intent?) {
            if (requestCode != REQUEST_LIVE) return
            val promise = pendingPromise ?: return
            pendingPromise = null
            if (resultCode != Activity.RESULT_OK) {
                promise.reject("TECHNIQUE_LIVE_CANCELLED", data?.getStringExtra(TechniqueLiveActivity.EXTRA_ERROR) ?: "Live-анализ отменён")
                return
            }
            val path = data?.getStringExtra(TechniqueLiveActivity.EXTRA_RESULT_PATH)
            if (path.isNullOrBlank()) {
                promise.reject("TECHNIQUE_LIVE_RESULT_MISSING", "Результат live-анализа отсутствует")
                return
            }
            val file = File(path)
            if (!file.exists()) {
                promise.reject("TECHNIQUE_LIVE_RESULT_MISSING", "Файл результата live-анализа отсутствует")
                return
            }
            try {
                promise.resolve(file.readText())
            } catch (error: Throwable) {
                promise.reject("TECHNIQUE_LIVE_RESULT_READ_FAILED", error)
            } finally {
                file.delete()
            }
        }
    }

    init { reactApplicationContext.addActivityEventListener(listener) }

    override fun getName() = NAME

    override fun startLiveSession(exerciseKey: String, maxDurationSeconds: Double, promise: Promise) {
        if (pendingPromise != null) {
            promise.reject("TECHNIQUE_LIVE_BUSY", "Live-анализ уже открыт")
            return
        }
        val activity = reactApplicationContext.currentActivity ?: run {
            promise.reject("TECHNIQUE_LIVE_NO_ACTIVITY", "Нет активного Android Activity")
            return
        }
        val key = exerciseKey.trim()
        if (key !in SUPPORTED) {
            promise.reject("TECHNIQUE_LIVE_UNSUPPORTED", "Упражнение пока не поддерживается live-анализом")
            return
        }
        pendingPromise = promise
        activity.startActivityForResult(
            Intent(activity, TechniqueLiveActivity::class.java).apply {
                putExtra(TechniqueLiveActivity.EXTRA_EXERCISE_KEY, key)
                putExtra(TechniqueLiveActivity.EXTRA_MAX_DURATION, maxDurationSeconds.toInt().coerceIn(10, 60))
            },
            REQUEST_LIVE,
        )
    }

    companion object {
        const val NAME = "TechniqueLive"
        private const val REQUEST_LIVE = 4331
        private val SUPPORTED = setOf("squat", "biceps_curl", "push_up", "lunge", "shoulder_press")
    }
}

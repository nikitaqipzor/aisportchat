package com.aifitnessos.nativebridge

import com.aifitnessos.BuildConfig
import com.aifitnessos.specs.NativeAIFitnessConfigSpec
import com.facebook.react.bridge.ReactApplicationContext
import com.facebook.react.bridge.Promise
import java.io.File

class AIFitnessConfigModule(reactContext: ReactApplicationContext) : NativeAIFitnessConfigSpec(reactContext) {
    override fun getName() = NAME
    override fun getApiBaseUrl(): String = BuildConfig.API_BASE_URL

    override fun clearPrivateTempFiles(promise: Promise) {
        try {
            val dirs = listOf("body_scan_camera", "technique_video", "technique_live")
            val deleted = dirs.map { name ->
                val path = File(reactApplicationContext.cacheDir, name)
                !path.exists() || path.deleteRecursively()
            }.all { it }
            promise.resolve(deleted)
        } catch (error: Throwable) {
            promise.reject("PRIVATE_CACHE_CLEAR_FAILED", error)
        }
    }

    companion object {
        const val NAME = "AIFitnessConfig"
    }
}

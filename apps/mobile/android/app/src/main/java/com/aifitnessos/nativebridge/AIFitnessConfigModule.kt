package com.aifitnessos.nativebridge

import com.aifitnessos.BuildConfig
import com.aifitnessos.specs.NativeAIFitnessConfigSpec
import com.facebook.react.bridge.ReactApplicationContext

class AIFitnessConfigModule(reactContext: ReactApplicationContext) : NativeAIFitnessConfigSpec(reactContext) {
    override fun getName() = NAME
    override fun getApiBaseUrl(): String = BuildConfig.API_BASE_URL

    companion object {
        const val NAME = "AIFitnessConfig"
    }
}

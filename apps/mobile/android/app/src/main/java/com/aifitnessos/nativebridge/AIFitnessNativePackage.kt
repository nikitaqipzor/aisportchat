package com.aifitnessos.nativebridge

import com.aifitnessos.foodphoto.FoodPhotoPickerModule
import com.aifitnessos.healthconnect.HealthConnectModule
import com.aifitnessos.bodyphoto.BodyPhotoCaptureModule
import com.aifitnessos.resttimer.RestTimerNotificationsModule
import com.aifitnessos.technique.TechniqueVideoModule
import com.aifitnessos.technique.TechniqueLiveModule
import com.facebook.react.BaseReactPackage
import com.facebook.react.bridge.NativeModule
import com.facebook.react.bridge.ReactApplicationContext
import com.facebook.react.module.model.ReactModuleInfo
import com.facebook.react.module.model.ReactModuleInfoProvider

class AIFitnessNativePackage : BaseReactPackage() {
    override fun getModule(name: String, reactContext: ReactApplicationContext): NativeModule? =
        when (name) {
            RestTimerNotificationsModule.NAME -> RestTimerNotificationsModule(reactContext)
            FoodPhotoPickerModule.NAME -> FoodPhotoPickerModule(reactContext)
            AIFitnessConfigModule.NAME -> AIFitnessConfigModule(reactContext)
            BodyPhotoCaptureModule.NAME -> BodyPhotoCaptureModule(reactContext)
            TechniqueVideoModule.NAME -> TechniqueVideoModule(reactContext)
            TechniqueLiveModule.NAME -> TechniqueLiveModule(reactContext)
            HealthConnectModule.NAME -> HealthConnectModule(reactContext)
            else -> null
        }

    override fun getReactModuleInfoProvider() = ReactModuleInfoProvider {
        mapOf(
            RestTimerNotificationsModule.NAME to ReactModuleInfo(
                RestTimerNotificationsModule.NAME,
                RestTimerNotificationsModule::class.java.name,
                false, false, false, true,
            ),
            FoodPhotoPickerModule.NAME to ReactModuleInfo(
                FoodPhotoPickerModule.NAME,
                FoodPhotoPickerModule::class.java.name,
                false, false, false, true,
            ),

            BodyPhotoCaptureModule.NAME to ReactModuleInfo(
                BodyPhotoCaptureModule.NAME,
                BodyPhotoCaptureModule::class.java.name,
                false, false, false, true,
            ),
            TechniqueVideoModule.NAME to ReactModuleInfo(
                TechniqueVideoModule.NAME,
                TechniqueVideoModule::class.java.name,
                false, false, false, true,
            ),
            TechniqueLiveModule.NAME to ReactModuleInfo(
                TechniqueLiveModule.NAME,
                TechniqueLiveModule::class.java.name,
                false, false, false, true,
            ),
            HealthConnectModule.NAME to ReactModuleInfo(
                HealthConnectModule.NAME,
                HealthConnectModule::class.java.name,
                false, false, false, true,
            ),
            AIFitnessConfigModule.NAME to ReactModuleInfo(
                AIFitnessConfigModule.NAME,
                AIFitnessConfigModule::class.java.name,
                false, false, false, true,
            ),
        )
    }
}

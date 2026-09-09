package com.aifitnessos.healthconnect

import android.app.Activity
import android.content.Intent
import androidx.health.connect.client.HealthConnectClient
import androidx.health.connect.client.HealthConnectFeatures
import androidx.health.connect.client.PermissionController
import com.aifitnessos.specs.NativeHealthConnectSpec
import com.facebook.react.bridge.ActivityEventListener
import com.facebook.react.bridge.BaseActivityEventListener
import com.facebook.react.bridge.Promise
import com.facebook.react.bridge.ReactApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch
import org.json.JSONArray
import org.json.JSONObject

class HealthConnectModule(reactContext: ReactApplicationContext) : NativeHealthConnectSpec(reactContext) {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var pendingPermissionPromise: Promise? = null
    private var pendingPermissionRequestCode: Int? = null
    private val permissionContract = PermissionController.createRequestPermissionResultContract()

    private val activityListener: ActivityEventListener = object : BaseActivityEventListener() {
        override fun onActivityResult(activity: Activity, requestCode: Int, resultCode: Int, data: Intent?) {
            if (requestCode != pendingPermissionRequestCode) return
            val promise = pendingPermissionPromise ?: return
            pendingPermissionPromise = null
            pendingPermissionRequestCode = null
            permissionContract.parseResult(resultCode, data)
            scope.launch {
                try { promise.resolve(statusJson().toString()) }
                catch (error: Throwable) { promise.reject("HEALTH_CONNECT_PERMISSION_STATUS_FAILED", error) }
            }
        }
    }

    init { reactApplicationContext.addActivityEventListener(activityListener) }
    override fun getName() = NAME

    override fun invalidate() {
        reactApplicationContext.removeActivityEventListener(activityListener)
        scope.cancel()
        super.invalidate()
    }

    override fun getStatus(promise: Promise) {
        scope.launch {
            try { promise.resolve(statusJson().toString()) }
            catch (error: Throwable) { promise.reject("HEALTH_CONNECT_STATUS_FAILED", error) }
        }
    }

    override fun requestPermissions(promise: Promise) = requestPermissionSet(HealthConnectReader.REQUIRED_PERMISSIONS, REQUEST_HEALTH_PERMISSIONS, promise)

    override fun requestBackgroundPermission(promise: Promise) = requestPermissionSet(setOf(HealthConnectReader.BACKGROUND_PERMISSION), REQUEST_BACKGROUND_PERMISSION, promise)

    private fun requestPermissionSet(permissions: Set<String>, requestCode: Int, promise: Promise) {
        if (pendingPermissionPromise != null) {
            promise.reject("HEALTH_CONNECT_PERMISSION_BUSY", "Health Connect permission request is already open")
            return
        }
        val activity = reactApplicationContext.currentActivity ?: run {
            promise.reject("HEALTH_CONNECT_NO_ACTIVITY", "No foreground Android Activity")
            return
        }
        if (HealthConnectClient.getSdkStatus(reactApplicationContext) != HealthConnectClient.SDK_AVAILABLE) {
            promise.reject("HEALTH_CONNECT_UNAVAILABLE", "Health Connect is not available on this device")
            return
        }
        pendingPermissionPromise = promise
        pendingPermissionRequestCode = requestCode
        val intent = permissionContract.createIntent(activity, permissions)
        activity.startActivityForResult(intent, requestCode)
    }

    override fun openSettings(promise: Promise) {
        val activity = reactApplicationContext.currentActivity ?: run { promise.resolve(false); return }
        try {
            activity.startActivity(Intent(HealthConnectClient.ACTION_HEALTH_CONNECT_SETTINGS))
            promise.resolve(true)
        } catch (_: Throwable) { promise.resolve(false) }
    }

    override fun readDailySnapshot(date: String, preferredSourcePackage: String, promise: Promise) {
        scope.launch {
            try { promise.resolve(HealthConnectReader.readSnapshot(reactApplicationContext, date, preferredSourcePackage.trim()).toString()) }
            catch (error: SecurityException) { promise.reject("HEALTH_CONNECT_PERMISSION_REQUIRED", error) }
            catch (error: Throwable) { promise.reject("HEALTH_CONNECT_READ_FAILED", error) }
        }
    }

    override fun setPassiveSyncEnabled(ownerUserId: String, enabled: Boolean, promise: Promise) {
        scope.launch {
            try {
                val owner = ownerUserId.trim()
                if (owner.isBlank()) error("Passive Health Sync requires an authenticated owner")
                if (enabled) {
                    val hc = HealthConnectReader.client(reactApplicationContext)
                    val available = hc.features.getFeatureStatus(HealthConnectFeatures.FEATURE_READ_HEALTH_DATA_IN_BACKGROUND) == HealthConnectFeatures.FEATURE_STATUS_AVAILABLE
                    val granted = hc.permissionController.getGrantedPermissions()
                    if (!available) error("Background Health Connect read is unavailable")
                    if (HealthConnectReader.BACKGROUND_PERMISSION !in granted) throw SecurityException("Background Health Connect permission is required")
                }
                HealthPassiveSyncScheduler.setEnabled(reactApplicationContext, owner, enabled)
                promise.resolve(statusJson().toString())
            } catch (error: SecurityException) { promise.reject("HEALTH_CONNECT_BACKGROUND_PERMISSION_REQUIRED", error) }
            catch (error: Throwable) { promise.reject("HEALTH_CONNECT_PASSIVE_SYNC_FAILED", error) }
        }
    }

    override fun getPendingSnapshots(ownerUserId: String, promise: Promise) {
        scope.launch {
            try { promise.resolve(HealthPassiveSyncStore.pending(reactApplicationContext, ownerUserId.trim()).toString()) }
            catch (error: Throwable) { promise.reject("HEALTH_CONNECT_PENDING_READ_FAILED", error) }
        }
    }

    override fun ackPendingSnapshot(ownerUserId: String, date: String, promise: Promise) {
        scope.launch {
            try { promise.resolve(HealthPassiveSyncStore.ack(reactApplicationContext, ownerUserId.trim(), date.trim())) }
            catch (error: Throwable) { promise.reject("HEALTH_CONNECT_PENDING_ACK_FAILED", error) }
        }
    }

    override fun clearPassiveSync(ownerUserId: String, promise: Promise) {
        scope.launch {
            try {
                val owner = ownerUserId.trim()
                if (owner.isNotBlank()) HealthPassiveSyncScheduler.disableAndClear(reactApplicationContext, owner)
                promise.resolve(true)
            } catch (error: Throwable) { promise.reject("HEALTH_CONNECT_PASSIVE_CLEAR_FAILED", error) }
        }
    }

    override fun bindSessionOwner(ownerUserId: String, promise: Promise) {
        scope.launch {
            try {
                val owner = ownerUserId.trim()
                if (owner.isNotBlank()) HealthPassiveSyncScheduler.bindAuthenticatedOwner(reactApplicationContext, owner)
                promise.resolve(true)
            } catch (error: Throwable) { promise.reject("HEALTH_CONNECT_OWNER_BIND_FAILED", error) }
        }
    }

    private suspend fun statusJson(): JSONObject {
        val sdk = HealthConnectClient.getSdkStatus(reactApplicationContext)
        val hc = if (sdk == HealthConnectClient.SDK_AVAILABLE) HealthConnectReader.client(reactApplicationContext) else null
        val granted = hc?.permissionController?.getGrantedPermissions() ?: emptySet()
        val missing = HealthConnectReader.REQUIRED_PERMISSIONS - granted
        val backgroundAvailable = hc?.features?.getFeatureStatus(HealthConnectFeatures.FEATURE_READ_HEALTH_DATA_IN_BACKGROUND) == HealthConnectFeatures.FEATURE_STATUS_AVAILABLE
        val passive = HealthPassiveSyncStore.status(reactApplicationContext)
        return JSONObject().apply {
            put("sdk_status", when (sdk) {
                HealthConnectClient.SDK_AVAILABLE -> "available"
                HealthConnectClient.SDK_UNAVAILABLE_PROVIDER_UPDATE_REQUIRED -> "update_required"
                else -> "unavailable"
            })
            put("permissions_granted", missing.isEmpty() && sdk == HealthConnectClient.SDK_AVAILABLE)
            put("granted_permissions", JSONArray(granted.sorted()))
            put("missing_permissions", JSONArray(missing.sorted()))
            put("mi_fitness_installed", packageInstalled(HealthConnectReader.MI_FITNESS_PACKAGE))
            put("background_read_available", backgroundAvailable)
            put("background_read_granted", HealthConnectReader.BACKGROUND_PERMISSION in granted)
            put("passive_sync", passive)
        }
    }

    private fun packageInstalled(packageName: String): Boolean = try {
        reactApplicationContext.packageManager.getPackageInfo(packageName, 0)
        true
    } catch (_: Throwable) { false }

    companion object {
        const val NAME = "HealthConnect"
        private const val REQUEST_HEALTH_PERMISSIONS = 4511
        private const val REQUEST_BACKGROUND_PERMISSION = 4512
    }
}

package com.aifitnessos.healthconnect

import android.content.Context
import androidx.health.connect.client.HealthConnectClient
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import java.time.LocalDate

class HealthPassiveSyncWorker(appContext: Context, workerParams: WorkerParameters) : CoroutineWorker(appContext, workerParams) {
    override suspend fun doWork(): Result {
        if (!HealthPassiveSyncStore.enabled(applicationContext)) return Result.success()
        val ownerUserId = HealthPassiveSyncStore.activeOwner(applicationContext) ?: return Result.success()
        if (HealthConnectClient.getSdkStatus(applicationContext) != HealthConnectClient.SDK_AVAILABLE) {
            HealthPassiveSyncStore.markError(applicationContext, "Health Connect unavailable")
            return Result.success()
        }
        return try {
            val client = HealthConnectReader.client(applicationContext)
            val granted = client.permissionController.getGrantedPermissions()
            if (HealthConnectReader.BACKGROUND_PERMISSION !in granted) {
                HealthPassiveSyncStore.markError(applicationContext, "Background Health Connect permission missing")
                return Result.success()
            }
            val today = LocalDate.now()
            listOf(today.minusDays(1), today).forEach { day ->
                val snapshot = HealthConnectReader.readSnapshot(applicationContext, day.toString(), HealthConnectReader.MI_FITNESS_PACKAGE)
                if ((snapshot.optJSONArray("data_types")?.length() ?: 0) > 0) HealthPassiveSyncStore.put(applicationContext, ownerUserId, snapshot)
            }
            HealthPassiveSyncStore.markSuccess(applicationContext)
            Result.success()
        } catch (error: SecurityException) {
            HealthPassiveSyncStore.markError(applicationContext, error.message ?: "Health permission error")
            Result.success()
        } catch (error: Throwable) {
            HealthPassiveSyncStore.markError(applicationContext, error.message ?: "Background sync failed")
            Result.retry()
        }
    }
}

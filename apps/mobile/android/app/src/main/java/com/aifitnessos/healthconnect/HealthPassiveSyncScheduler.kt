package com.aifitnessos.healthconnect

import android.content.Context
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import java.util.concurrent.TimeUnit

object HealthPassiveSyncScheduler {
    private const val UNIQUE_WORK = "ai_fitness_health_passive_sync"

    fun setEnabled(context: Context, ownerUserId: String, enabled: Boolean) {
        HealthPassiveSyncStore.setEnabled(context, ownerUserId, enabled)
        val manager = WorkManager.getInstance(context)
        if (!enabled) {
            manager.cancelUniqueWork(UNIQUE_WORK)
            return
        }
        val request = PeriodicWorkRequestBuilder<HealthPassiveSyncWorker>(1, TimeUnit.HOURS).build()
        manager.enqueueUniquePeriodicWork(UNIQUE_WORK, ExistingPeriodicWorkPolicy.UPDATE, request)
    }

    fun bindAuthenticatedOwner(context: Context, ownerUserId: String) {
        val owner = ownerUserId.trim()
        if (owner.isBlank()) return
        val active = HealthPassiveSyncStore.activeOwner(context) ?: return
        if (active == owner) return
        WorkManager.getInstance(context).cancelUniqueWork(UNIQUE_WORK)
        HealthPassiveSyncStore.suspendActiveOwner(context)
    }

    fun disableAndClear(context: Context, ownerUserId: String) {
        WorkManager.getInstance(context).cancelUniqueWork(UNIQUE_WORK)
        HealthPassiveSyncStore.clearOwner(context, ownerUserId)
    }
}

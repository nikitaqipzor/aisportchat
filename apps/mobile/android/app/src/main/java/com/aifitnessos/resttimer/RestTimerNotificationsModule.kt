package com.aifitnessos.resttimer

import android.app.AlarmManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import com.aifitnessos.specs.NativeRestTimerNotificationsSpec
import com.facebook.react.bridge.Promise
import com.facebook.react.bridge.ReactApplicationContext

class RestTimerNotificationsModule(reactContext: ReactApplicationContext) : NativeRestTimerNotificationsSpec(reactContext) {
    override fun getName() = NAME

    override fun schedule(seconds: Double, exerciseName: String, promise: Promise) {
        try {
            val delayMs = (seconds.coerceAtLeast(1.0) * 1000).toLong()
            val intent = Intent(reactApplicationContext, RestTimerReceiver::class.java).apply {
                putExtra(RestTimerReceiver.EXTRA_EXERCISE, exerciseName)
            }
            val pending = PendingIntent.getBroadcast(
                reactApplicationContext,
                REQUEST_CODE,
                intent,
                PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
            )
            val alarm = reactApplicationContext.getSystemService(Context.ALARM_SERVICE) as AlarmManager
            val triggerAt = System.currentTimeMillis() + delayMs
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                alarm.setAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, triggerAt, pending)
            } else {
                alarm.set(AlarmManager.RTC_WAKEUP, triggerAt, pending)
            }
            promise.resolve(true)
        } catch (error: Throwable) {
            promise.reject("REST_TIMER_SCHEDULE_FAILED", error)
        }
    }

    override fun cancel(promise: Promise) {
        try {
            val intent = Intent(reactApplicationContext, RestTimerReceiver::class.java)
            val pending = PendingIntent.getBroadcast(
                reactApplicationContext,
                REQUEST_CODE,
                intent,
                PendingIntent.FLAG_NO_CREATE or PendingIntent.FLAG_IMMUTABLE,
            )
            if (pending != null) {
                val alarm = reactApplicationContext.getSystemService(Context.ALARM_SERVICE) as AlarmManager
                alarm.cancel(pending)
                pending.cancel()
            }
            promise.resolve(true)
        } catch (error: Throwable) {
            promise.reject("REST_TIMER_CANCEL_FAILED", error)
        }
    }

    companion object {
        const val NAME = "RestTimerNotifications"
        private const val REQUEST_CODE = 4107
    }
}

package com.aifitnessos.resttimer

import android.Manifest
import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat

class RestTimerReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        createChannel(context)
        if (Build.VERSION.SDK_INT >= 33 &&
            context.checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED
        ) return

        val exercise = intent.getStringExtra(EXTRA_EXERCISE).orEmpty()
        val text = if (exercise.isBlank()) {
            "Можно начинать следующий подход."
        } else {
            "$exercise: можно начинать следующий подход."
        }

        val notification = NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(android.R.drawable.ic_popup_reminder)
            .setContentTitle("Отдых завершён")
            .setContentText(text)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setAutoCancel(true)
            .build()
        NotificationManagerCompat.from(context).notify(NOTIFICATION_ID, notification)
    }

    private fun createChannel(context: Context) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val manager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        manager.createNotificationChannel(
            NotificationChannel(CHANNEL_ID, "Таймер отдыха", NotificationManager.IMPORTANCE_HIGH).apply {
                description = "Уведомления о завершении отдыха между подходами"
            },
        )
    }

    companion object {
        const val EXTRA_EXERCISE = "exercise_name"
        private const val CHANNEL_ID = "rest_timer"
        private const val NOTIFICATION_ID = 4107
    }
}

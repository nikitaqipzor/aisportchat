package com.aifitnessos.healthconnect

import android.app.Activity
import android.os.Bundle
import android.text.method.LinkMovementMethod
import android.view.ViewGroup
import android.widget.ScrollView
import android.widget.TextView

class PermissionsRationaleActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val density = resources.displayMetrics.density
        val text = TextView(this).apply {
            textSize = 17f
            setPadding((24*density).toInt(), (32*density).toInt(), (24*density).toInt(), (32*density).toInt())
            movementMethod = LinkMovementMethod.getInstance()
            text = """AI Fitness OS использует разрешённые вами данные Health Connect только для фитнес-аналитики и Recovery/Readiness.\n\nМы запрашиваем чтение шагов, дистанции, активных калорий, сна, тренировочных сессий и пульса во время тренировок.\n\nДанные не изменяются в Health Connect. Доступ можно отозвать в любой момент в настройках Health Connect. Приложение не использует эти показатели для медицинской диагностики."""
        }
        setContentView(ScrollView(this).apply { addView(text, ViewGroup.LayoutParams(ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.WRAP_CONTENT)) })
    }
}

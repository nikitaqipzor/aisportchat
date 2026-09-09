package com.aifitnessos.healthconnect

import android.content.Context
import androidx.health.connect.client.HealthConnectClient
import androidx.health.connect.client.permission.HealthPermission
import androidx.health.connect.client.records.ActiveCaloriesBurnedRecord
import androidx.health.connect.client.records.DistanceRecord
import androidx.health.connect.client.records.ExerciseSessionRecord
import androidx.health.connect.client.records.HeartRateRecord
import androidx.health.connect.client.records.SleepSessionRecord
import androidx.health.connect.client.records.StepsRecord
import androidx.health.connect.client.records.metadata.DataOrigin
import androidx.health.connect.client.request.AggregateRequest
import androidx.health.connect.client.request.ReadRecordsRequest
import androidx.health.connect.client.time.TimeRangeFilter
import org.json.JSONArray
import org.json.JSONObject
import java.time.Duration
import java.time.LocalDate
import java.time.ZoneId
import java.time.format.DateTimeParseException
import kotlin.math.round

object HealthConnectReader {
    const val MI_FITNESS_PACKAGE = "com.xiaomi.wearable"
    const val BACKGROUND_PERMISSION = "android.permission.health.READ_HEALTH_DATA_IN_BACKGROUND"

    val REQUIRED_PERMISSIONS = setOf(
        HealthPermission.getReadPermission(StepsRecord::class),
        HealthPermission.getReadPermission(DistanceRecord::class),
        HealthPermission.getReadPermission(ActiveCaloriesBurnedRecord::class),
        HealthPermission.getReadPermission(SleepSessionRecord::class),
        HealthPermission.getReadPermission(ExerciseSessionRecord::class),
        HealthPermission.getReadPermission(HeartRateRecord::class),
    )

    fun client(context: Context): HealthConnectClient {
        if (HealthConnectClient.getSdkStatus(context) != HealthConnectClient.SDK_AVAILABLE) error("Health Connect is unavailable")
        return HealthConnectClient.getOrCreate(context)
    }

    suspend fun readSnapshot(context: Context, dateRaw: String, preferredSourcePackage: String): JSONObject {
        val day = try { LocalDate.parse(dateRaw) } catch (_: DateTimeParseException) { error("date must be YYYY-MM-DD") }
        val hc = client(context)
        val granted = hc.permissionController.getGrantedPermissions()
        if (!granted.containsAll(REQUIRED_PERMISSIONS)) throw SecurityException("Health Connect permissions are incomplete")

        val zone = ZoneId.systemDefault()
        val start = day.atStartOfDay(zone).toInstant()
        val end = day.plusDays(1).atStartOfDay(zone).toInstant()
        val sourcePackage = preferredSourcePackage.ifBlank { MI_FITNESS_PACKAGE }
        val origins = setOf(DataOrigin(sourcePackage))
        val time = TimeRangeFilter.between(start, end)

        val aggregate = hc.aggregate(AggregateRequest(
            metrics = setOf(StepsRecord.COUNT_TOTAL, DistanceRecord.DISTANCE_TOTAL, ActiveCaloriesBurnedRecord.ACTIVE_CALORIES_TOTAL),
            timeRangeFilter = time,
            dataOriginFilter = origins,
        ))
        val steps = aggregate[StepsRecord.COUNT_TOTAL] ?: 0L
        val distance = aggregate[DistanceRecord.DISTANCE_TOTAL]?.inMeters ?: 0.0
        val calories = aggregate[ActiveCaloriesBurnedRecord.ACTIVE_CALORIES_TOTAL]?.inKilocalories ?: 0.0

        val sleepResponse = hc.readRecords(ReadRecordsRequest<SleepSessionRecord>(
            timeRangeFilter = TimeRangeFilter.between(start.minus(Duration.ofHours(12)), end),
            dataOriginFilter = origins,
        ))
        var sleepMinutes = 0L; var deep = 0L; var light = 0L; var rem = 0L; var awake = 0L
        sleepResponse.records.filter { !it.endTime.isBefore(start) && it.endTime.isBefore(end.plusSeconds(1)) }.forEach { record ->
            var stagedSleep = 0L
            record.stages.forEach { stage ->
                val minutes = Duration.between(stage.startTime, stage.endTime).toMinutes().coerceAtLeast(0)
                when (stage.stage) {
                    SleepSessionRecord.STAGE_TYPE_DEEP -> { deep += minutes; stagedSleep += minutes }
                    SleepSessionRecord.STAGE_TYPE_LIGHT -> { light += minutes; stagedSleep += minutes }
                    SleepSessionRecord.STAGE_TYPE_REM -> { rem += minutes; stagedSleep += minutes }
                    SleepSessionRecord.STAGE_TYPE_SLEEPING -> stagedSleep += minutes
                    SleepSessionRecord.STAGE_TYPE_AWAKE, SleepSessionRecord.STAGE_TYPE_AWAKE_IN_BED, SleepSessionRecord.STAGE_TYPE_OUT_OF_BED -> awake += minutes
                }
            }
            sleepMinutes += if (stagedSleep > 0) stagedSleep else Duration.between(record.startTime, record.endTime).toMinutes().coerceAtLeast(0)
        }

        val sessions = hc.readRecords(ReadRecordsRequest<ExerciseSessionRecord>(timeRangeFilter = time, dataOriginFilter = origins)).records
        val exerciseMinutes = sessions.sumOf { Duration.between(it.startTime, it.endTime).toMinutes().coerceAtLeast(0) }
        val heartRates = mutableListOf<Long>()
        sessions.forEach { session ->
            hc.readRecords(ReadRecordsRequest<HeartRateRecord>(
                timeRangeFilter = TimeRangeFilter.between(session.startTime, session.endTime), dataOriginFilter = origins,
            )).records.forEach { record -> record.samples.forEach { heartRates.add(it.beatsPerMinute) } }
        }

        val types = mutableListOf<String>()
        if (steps > 0) types += "steps"
        if (distance > 0) types += "distance"
        if (calories > 0) types += "active_calories"
        if (sleepMinutes > 0) types += "sleep"
        if (sessions.isNotEmpty()) types += "exercise"
        if (heartRates.isNotEmpty()) types += "heart_rate"

        return JSONObject().apply {
            put("date", day.toString()); put("provider", "health_connect"); put("source_package", sourcePackage)
            put("source_label", if (sourcePackage == MI_FITNESS_PACKAGE) "Mi Fitness (Xiaomi Wear)" else sourcePackage)
            put("steps", steps); put("distance_m", round2(distance)); put("active_calories_kcal", round2(calories))
            put("sleep_minutes", sleepMinutes.coerceAtMost(1440)); put("deep_sleep_minutes", deep.coerceAtMost(1440)); put("light_sleep_minutes", light.coerceAtMost(1440)); put("rem_sleep_minutes", rem.coerceAtMost(1440)); put("awake_minutes", awake.coerceAtMost(1440))
            put("exercise_minutes", exerciseMinutes.coerceAtMost(1440)); put("exercise_sessions", sessions.size)
            if (heartRates.isNotEmpty()) { put("exercise_heart_rate_avg", round2(heartRates.average())); put("exercise_heart_rate_max", heartRates.maxOrNull()) }
            put("data_types", JSONArray(types)); put("captured_at", java.time.Instant.now().toString())
        }
    }

    private fun round2(v: Double) = round(v * 100.0) / 100.0
}

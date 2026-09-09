package com.aifitnessos.technique

import kotlin.math.acos
import kotlin.math.hypot
import kotlin.math.min

class LiveRepCounter(private val exerciseKey: String) {
    private val bottom: Double
    private val top: Double
    private val startsAtBottom: Boolean
    private var ready = false
    private var inRep = false
    private var hitTurn = false
    private var startMs = 0L
    var count: Int = 0
        private set

    init {
        val thresholds = when (exerciseKey) {
            "squat" -> 105.0 to 155.0
            "biceps_curl" -> 75.0 to 145.0
            "push_up" -> 105.0 to 155.0
            "lunge" -> 110.0 to 150.0
            "shoulder_press" -> 105.0 to 150.0
            else -> 90.0 to 150.0
        }
        bottom = thresholds.first
        top = thresholds.second
        startsAtBottom = exerciseKey == "shoulder_press"
    }

    fun update(points: List<LivePosePoint>, timestampMs: Long): Int {
        val avg = movementAngle(points) ?: return count
        if (!inRep) {
            if (startsAtBottom) {
                if (avg <= bottom) ready = true
                if (ready && avg > bottom + 10.0) {
                    inRep = true; hitTurn = false; startMs = timestampMs
                } else return count
            } else {
                if (avg >= top) ready = true
                if (ready && avg < top - 10.0) {
                    inRep = true; hitTurn = false; startMs = timestampMs
                } else return count
            }
        }

        val complete = if (startsAtBottom) {
            if (avg >= top) hitTurn = true
            hitTurn && avg <= bottom
        } else {
            if (avg <= bottom) hitTurn = true
            hitTurn && avg >= top
        }
        if (complete) {
            val duration = timestampMs - startMs
            if (duration in 350L..10_000L) count += 1
            inRep = false
            hitTurn = false
            ready = true
        }
        return count
    }

    private fun movementAngle(points: List<LivePosePoint>): Double? {
        if (points.size < 33) return null
        return when (exerciseKey) {
            "squat" -> pairAverage(points, 23, 25, 27, 24, 26, 28)
            "biceps_curl", "push_up", "shoulder_press" -> pairAverage(points, 11, 13, 15, 12, 14, 16)
            "lunge" -> {
                val left = angle(points[23], points[25], points[27])
                val right = angle(points[24], points[26], points[28])
                if (left == null || right == null) null else min(left, right)
            }
            else -> null
        }
    }

    private fun pairAverage(points: List<LivePosePoint>, a1: Int, b1: Int, c1: Int, a2: Int, b2: Int, c2: Int): Double? {
        val left = angle(points[a1], points[b1], points[c1]) ?: return null
        val right = angle(points[a2], points[b2], points[c2]) ?: return null
        return (left + right) / 2.0
    }

    private fun angle(a: LivePosePoint, b: LivePosePoint, c: LivePosePoint): Double? {
        if (a.visibility < .35f || b.visibility < .35f || c.visibility < .35f) return null
        val v1x = (a.x - b.x).toDouble(); val v1y = (a.y - b.y).toDouble()
        val v2x = (c.x - b.x).toDouble(); val v2y = (c.y - b.y).toDouble()
        val denominator = hypot(v1x, v1y) * hypot(v2x, v2y)
        if (denominator < 1e-8) return null
        val cos = ((v1x * v2x + v1y * v2y) / denominator).coerceIn(-1.0, 1.0)
        return Math.toDegrees(acos(cos))
    }
}

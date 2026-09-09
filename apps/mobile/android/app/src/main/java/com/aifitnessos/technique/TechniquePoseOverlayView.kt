package com.aifitnessos.technique

import android.content.Context
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Paint
import android.view.View

data class LivePosePoint(val x: Float, val y: Float, val z: Float, val visibility: Float)

class TechniquePoseOverlayView(context: Context) : View(context) {
    private var points: List<LivePosePoint> = emptyList()
    private var imageWidth = 1
    private var imageHeight = 1
    private var mirrored = false

    private val linePaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
        color = Color.rgb(116, 255, 168)
        strokeWidth = 5f
        style = Paint.Style.STROKE
        strokeCap = Paint.Cap.ROUND
    }
    private val pointPaint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
        color = Color.WHITE
        style = Paint.Style.FILL
    }

    fun updatePose(next: List<LivePosePoint>, width: Int, height: Int, mirror: Boolean) {
        points = next
        imageWidth = width.coerceAtLeast(1)
        imageHeight = height.coerceAtLeast(1)
        mirrored = mirror
        postInvalidateOnAnimation()
    }

    fun clearPose() {
        points = emptyList()
        postInvalidateOnAnimation()
    }

    override fun onDraw(canvas: Canvas) {
        super.onDraw(canvas)
        if (points.size < 33) return
        val imageAspect = imageWidth.toFloat() / imageHeight.toFloat()
        val viewAspect = width.toFloat() / height.toFloat().coerceAtLeast(1f)
        val drawWidth: Float
        val drawHeight: Float
        val offsetX: Float
        val offsetY: Float
        if (viewAspect > imageAspect) {
            drawHeight = height.toFloat()
            drawWidth = drawHeight * imageAspect
            offsetX = (width - drawWidth) / 2f
            offsetY = 0f
        } else {
            drawWidth = width.toFloat()
            drawHeight = drawWidth / imageAspect
            offsetX = 0f
            offsetY = (height - drawHeight) / 2f
        }
        fun xy(index: Int): Pair<Float, Float>? {
            if (index !in points.indices) return null
            val p = points[index]
            if (p.visibility < 0.35f) return null
            val nx = if (mirrored) 1f - p.x else p.x
            return (offsetX + nx * drawWidth) to (offsetY + p.y * drawHeight)
        }
        for ((a, b) in CONNECTIONS) {
            val pa = xy(a) ?: continue
            val pb = xy(b) ?: continue
            canvas.drawLine(pa.first, pa.second, pb.first, pb.second, linePaint)
        }
        for (index in KEY_POINTS) {
            val p = xy(index) ?: continue
            canvas.drawCircle(p.first, p.second, 7f, pointPaint)
        }
    }

    companion object {
        private val KEY_POINTS = intArrayOf(0, 11, 12, 13, 14, 15, 16, 23, 24, 25, 26, 27, 28)
        private val CONNECTIONS = arrayOf(
            11 to 12, 11 to 13, 13 to 15, 12 to 14, 14 to 16,
            11 to 23, 12 to 24, 23 to 24, 23 to 25, 25 to 27, 24 to 26, 26 to 28,
        )
    }
}

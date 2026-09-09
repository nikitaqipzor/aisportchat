package com.aifitnessos.bodyphoto

import android.app.Activity
import android.content.Intent
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Matrix
import android.media.ExifInterface
import android.provider.MediaStore
import android.util.Base64
import androidx.core.content.FileProvider
import com.aifitnessos.specs.NativeBodyPhotoCaptureSpec
import com.facebook.react.bridge.ActivityEventListener
import com.facebook.react.bridge.BaseActivityEventListener
import com.facebook.react.bridge.Promise
import com.facebook.react.bridge.ReactApplicationContext
import java.io.ByteArrayOutputStream
import java.io.File

class BodyPhotoCaptureModule(reactContext: ReactApplicationContext) : NativeBodyPhotoCaptureSpec(reactContext) {
    private var pendingPromise: Promise? = null
    private var pendingFile: File? = null

    private val listener: ActivityEventListener = object : BaseActivityEventListener() {
        override fun onActivityResult(activity: Activity, requestCode: Int, resultCode: Int, data: Intent?) {
            if (requestCode != REQUEST_CAPTURE) return
            val promise = pendingPromise ?: return
            val file = pendingFile
            pendingPromise = null
            pendingFile = null
            if (resultCode != Activity.RESULT_OK || file == null || !file.exists()) {
                file?.delete()
                promise.reject("BODY_PHOTO_CANCELLED", "Съёмка отменена")
                return
            }
            try {
                promise.resolve(toDataURL(file))
            } catch (error: Throwable) {
                promise.reject("BODY_PHOTO_FAILED", error)
            } finally {
                file.delete()
            }
        }
    }

    init { reactApplicationContext.addActivityEventListener(listener) }

    override fun getName() = NAME

    override fun captureImage(promise: Promise) {
        if (pendingPromise != null) {
            promise.reject("BODY_PHOTO_BUSY", "Камера уже открыта")
            return
        }
        val activity = reactApplicationContext.currentActivity
        if (activity == null) {
            promise.reject("BODY_PHOTO_NO_ACTIVITY", "Нет активного Android Activity")
            return
        }
        val cameraIntent = Intent(MediaStore.ACTION_IMAGE_CAPTURE)
        if (cameraIntent.resolveActivity(reactApplicationContext.packageManager) == null) {
            promise.reject("BODY_PHOTO_NO_CAMERA", "Камера недоступна")
            return
        }
        val dir = File(reactApplicationContext.cacheDir, "body_scan_camera").apply { mkdirs() }
        val file = File.createTempFile("body_scan_", ".jpg", dir)
        val uri = FileProvider.getUriForFile(reactApplicationContext, "${reactApplicationContext.packageName}.fileprovider", file)
        cameraIntent.putExtra(MediaStore.EXTRA_OUTPUT, uri)
        cameraIntent.addFlags(Intent.FLAG_GRANT_WRITE_URI_PERMISSION or Intent.FLAG_GRANT_READ_URI_PERMISSION)
        pendingPromise = promise
        pendingFile = file
        activity.startActivityForResult(cameraIntent, REQUEST_CAPTURE)
    }

    private fun toDataURL(file: File): String {
        val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
        BitmapFactory.decodeFile(file.absolutePath, bounds)
        var sample = 1
        while (bounds.outWidth / sample > MAX_DIMENSION || bounds.outHeight / sample > MAX_DIMENSION) sample *= 2
        val bitmap = BitmapFactory.decodeFile(file.absolutePath, BitmapFactory.Options().apply { inSampleSize = sample })
            ?: error("Не удалось декодировать фотографию")
        val rotated = rotateForExif(bitmap, file)
        val out = ByteArrayOutputStream()
        if (!rotated.compress(Bitmap.CompressFormat.JPEG, JPEG_QUALITY, out)) error("Не удалось сжать фотографию")
        if (rotated !== bitmap) rotated.recycle()
        bitmap.recycle()
        val bytes = out.toByteArray()
        if (bytes.size > MAX_BYTES) error("Фотография после обработки превышает 8 МБ")
        return "data:image/jpeg;base64," + Base64.encodeToString(bytes, Base64.NO_WRAP)
    }

    private fun rotateForExif(bitmap: Bitmap, file: File): Bitmap {
        val orientation = try { ExifInterface(file.absolutePath).getAttributeInt(ExifInterface.TAG_ORIENTATION, ExifInterface.ORIENTATION_NORMAL) } catch (_: Throwable) { ExifInterface.ORIENTATION_NORMAL }
        val degrees = when (orientation) {
            ExifInterface.ORIENTATION_ROTATE_90 -> 90f
            ExifInterface.ORIENTATION_ROTATE_180 -> 180f
            ExifInterface.ORIENTATION_ROTATE_270 -> 270f
            else -> 0f
        }
        if (degrees == 0f) return bitmap
        return Bitmap.createBitmap(bitmap, 0, 0, bitmap.width, bitmap.height, Matrix().apply { postRotate(degrees) }, true)
    }

    companion object {
        const val NAME = "BodyPhotoCapture"
        private const val REQUEST_CAPTURE = 4317
        private const val MAX_DIMENSION = 2200
        private const val JPEG_QUALITY = 88
        private const val MAX_BYTES = 8 * 1024 * 1024
    }
}

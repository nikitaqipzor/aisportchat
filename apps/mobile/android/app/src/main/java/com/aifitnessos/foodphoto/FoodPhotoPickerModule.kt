package com.aifitnessos.foodphoto

import android.app.Activity
import android.content.Intent
import android.net.Uri
import android.provider.MediaStore
import android.util.Base64
import com.aifitnessos.specs.NativeFoodPhotoPickerSpec
import com.facebook.react.bridge.ActivityEventListener
import com.facebook.react.bridge.BaseActivityEventListener
import com.facebook.react.bridge.Promise
import com.facebook.react.bridge.ReactApplicationContext

class FoodPhotoPickerModule(reactContext: ReactApplicationContext) : NativeFoodPhotoPickerSpec(reactContext) {
    private var pendingPromise: Promise? = null

    private val activityListener: ActivityEventListener = object : BaseActivityEventListener() {
        override fun onActivityResult(activity: Activity, requestCode: Int, resultCode: Int, data: Intent?) {
            if (requestCode != REQUEST_PICK) return
            val promise = pendingPromise ?: return
            pendingPromise = null
            if (resultCode != Activity.RESULT_OK) {
                promise.reject("PHOTO_PICK_CANCELLED", "Photo selection was cancelled")
                return
            }
            val uri = data?.data
            if (uri == null) {
                promise.reject("PHOTO_PICK_EMPTY", "No image was returned")
                return
            }
            try {
                promise.resolve(readDataURL(uri))
            } catch (error: FoodPhotoException) {
                promise.reject(error.code, error.message, error)
            } catch (error: Throwable) {
                promise.reject("PHOTO_PICK_FAILED", error)
            }
        }
    }

    init { reactApplicationContext.addActivityEventListener(activityListener) }

    override fun getName() = NAME

    override fun pickImage(promise: Promise) {
        if (pendingPromise != null) {
            promise.reject("PHOTO_PICK_BUSY", "Photo picker is already open")
            return
        }
        val activity = reactApplicationContext.currentActivity
        if (activity == null) {
            promise.reject("PHOTO_PICK_NO_ACTIVITY", "No foreground activity")
            return
        }
        pendingPromise = promise
        val intent = Intent(Intent.ACTION_PICK, MediaStore.Images.Media.EXTERNAL_CONTENT_URI).apply {
            type = "image/*"
        }
        activity.startActivityForResult(intent, REQUEST_PICK)
    }

    private fun readDataURL(uri: Uri): String {
        val resolver = reactApplicationContext.contentResolver
        val mime = resolver.getType(uri)?.lowercase() ?: "image/jpeg"
        if (mime !in SUPPORTED_MIME_TYPES) {
            throw FoodPhotoException("PHOTO_PICK_UNSUPPORTED", "Unsupported image format")
        }
        val input = resolver.openInputStream(uri)
            ?: throw FoodPhotoException("PHOTO_PICK_EMPTY", "Unable to open selected image")
        val bytes = input.use { stream ->
            val output = java.io.ByteArrayOutputStream()
            val buffer = ByteArray(16 * 1024)
            var total = 0
            while (true) {
                val read = stream.read(buffer)
                if (read < 0) break
                total += read
                if (total > MAX_BYTES) {
                    throw FoodPhotoException("PHOTO_PICK_TOO_LARGE", "Image exceeds 5 MB limit")
                }
                output.write(buffer, 0, read)
            }
            output.toByteArray()
        }
        if (bytes.isEmpty()) throw FoodPhotoException("PHOTO_PICK_EMPTY", "Selected image is empty")
        return "data:$mime;base64," + Base64.encodeToString(bytes, Base64.NO_WRAP)
    }

    companion object {
        const val NAME = "FoodPhotoPicker"
        private const val REQUEST_PICK = 4212
        private const val MAX_BYTES = 5 * 1024 * 1024
        private val SUPPORTED_MIME_TYPES = setOf("image/jpeg", "image/png", "image/webp")
    }
}

private class FoodPhotoException(val code: String, message: String) : Exception(message)

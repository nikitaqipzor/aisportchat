package com.aifitnessos.healthconnect

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import org.json.JSONArray
import org.json.JSONObject
import java.nio.charset.StandardCharsets
import java.security.KeyStore
import java.time.Instant
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

object HealthPassiveSyncStore {
    private const val PREFS = "health_passive_sync"
    private const val KEY_PAYLOAD = "pending_encrypted_v2"
    private const val LEGACY_KEY_PAYLOAD = "pending_encrypted"
    private const val KEY_ENABLED = "enabled"
    private const val KEY_ACTIVE_OWNER = "active_owner"
    private const val KEY_LAST_SUCCESS = "last_success"
    private const val KEY_LAST_ERROR = "last_error"
    private const val KEY_ALIAS = "ai_fitness_health_cache_v1"

    private fun prefs(context: Context) = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
    private fun normalizedOwner(ownerUserId: String) = ownerUserId.trim()

    fun activeOwner(context: Context): String? = prefs(context).getString(KEY_ACTIVE_OWNER, null)?.takeIf { it.isNotBlank() }

    fun enabled(context: Context): Boolean = prefs(context).getBoolean(KEY_ENABLED, false) && activeOwner(context) != null

    fun setEnabled(context: Context, ownerUserId: String, enabled: Boolean) {
        val owner = normalizedOwner(ownerUserId)
        require(owner.isNotBlank()) { "Passive Health Sync requires an authenticated owner" }
        val edit = prefs(context).edit()
        if (enabled) {
            edit.putBoolean(KEY_ENABLED, true).putString(KEY_ACTIVE_OWNER, owner)
        } else if (activeOwner(context) == owner) {
            edit.putBoolean(KEY_ENABLED, false).remove(KEY_ACTIVE_OWNER)
        }
        edit.apply()
    }

    fun status(context: Context): JSONObject {
        discardUnscopedLegacyCache(context)
        val owner = activeOwner(context)
        return JSONObject().apply {
            put("enabled", enabled(context))
            put("owner_bound", owner != null)
            put("last_success_at", prefs(context).getString(KEY_LAST_SUCCESS, null))
            put("last_error", prefs(context).getString(KEY_LAST_ERROR, "") ?: "")
            put("pending_snapshots", if (owner == null) 0 else pending(context, owner).length())
        }
    }

    fun suspendActiveOwner(context: Context) {
        prefs(context).edit().putBoolean(KEY_ENABLED, false).remove(KEY_ACTIVE_OWNER).apply()
    }

    fun markSuccess(context: Context) = prefs(context).edit()
        .putString(KEY_LAST_SUCCESS, Instant.now().toString()).putString(KEY_LAST_ERROR, "").apply()

    fun markError(context: Context, message: String) = prefs(context).edit()
        .putString(KEY_LAST_ERROR, message.take(240)).apply()

    @Synchronized fun put(context: Context, ownerUserId: String, snapshot: JSONObject) {
        val owner = normalizedOwner(ownerUserId)
        val date = snapshot.optString("date")
        if (owner.isBlank() || date.isBlank()) return
        val root = readRoot(context)
        val ownerMap = root.optJSONObject(owner) ?: JSONObject()
        ownerMap.put(date, snapshot)
        root.put(owner, ownerMap)
        writeRoot(context, root)
    }

    @Synchronized fun pending(context: Context, ownerUserId: String): JSONArray {
        val owner = normalizedOwner(ownerUserId)
        if (owner.isBlank()) return JSONArray()
        val map = readRoot(context).optJSONObject(owner) ?: return JSONArray()
        val out = JSONArray()
        map.keys().asSequence().toList().sorted().forEach { key -> map.optJSONObject(key)?.let { out.put(it) } }
        return out
    }

    @Synchronized fun ack(context: Context, ownerUserId: String, date: String): Boolean {
        val owner = normalizedOwner(ownerUserId)
        if (owner.isBlank()) return false
        val root = readRoot(context)
        val map = root.optJSONObject(owner) ?: return false
        if (!map.has(date)) return false
        map.remove(date)
        if (map.length() == 0) root.remove(owner) else root.put(owner, map)
        writeRoot(context, root)
        return true
    }

    @Synchronized fun clearOwner(context: Context, ownerUserId: String) {
        val owner = normalizedOwner(ownerUserId)
        if (owner.isBlank()) return
        val root = readRoot(context)
        if (root.has(owner)) {
            root.remove(owner)
            writeRoot(context, root)
        }
        if (activeOwner(context) == owner) {
            prefs(context).edit().putBoolean(KEY_ENABLED, false).remove(KEY_ACTIVE_OWNER).apply()
        }
    }

    private fun discardUnscopedLegacyCache(context: Context) {
        // v1 snapshots had no account owner. They cannot be safely attributed after upgrade,
        // so privacy wins over delivery and the unscoped queue is discarded once.
        if (prefs(context).contains(LEGACY_KEY_PAYLOAD)) prefs(context).edit().remove(LEGACY_KEY_PAYLOAD).apply()
    }

    private fun readRoot(context: Context): JSONObject {
        discardUnscopedLegacyCache(context)
        val raw = prefs(context).getString(KEY_PAYLOAD, null) ?: return JSONObject()
        return try { JSONObject(String(decrypt(raw), StandardCharsets.UTF_8)) } catch (_: Throwable) { JSONObject() }
    }

    private fun writeRoot(context: Context, value: JSONObject) {
        val editor = prefs(context).edit()
        if (value.length() == 0) editor.remove(KEY_PAYLOAD)
        else editor.putString(KEY_PAYLOAD, encrypt(value.toString().toByteArray(StandardCharsets.UTF_8)))
        editor.apply()
    }

    private fun secretKey(): SecretKey {
        val store = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
        (store.getKey(KEY_ALIAS, null) as? SecretKey)?.let { return it }
        val generator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore")
        generator.init(KeyGenParameterSpec.Builder(KEY_ALIAS, KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT)
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM).setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE).build())
        return generator.generateKey()
    }

    private fun encrypt(bytes: ByteArray): String {
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, secretKey())
        val iv = cipher.iv
        val encrypted = cipher.doFinal(bytes)
        val packet = ByteArray(1 + iv.size + encrypted.size)
        packet[0] = iv.size.toByte()
        System.arraycopy(iv, 0, packet, 1, iv.size)
        System.arraycopy(encrypted, 0, packet, 1 + iv.size, encrypted.size)
        return Base64.encodeToString(packet, Base64.NO_WRAP)
    }

    private fun decrypt(raw: String): ByteArray {
        val packet = Base64.decode(raw, Base64.NO_WRAP)
        require(packet.isNotEmpty()) { "Encrypted health cache is empty" }
        val ivSize = packet[0].toInt() and 0xff
        require(ivSize > 0 && packet.size > 1 + ivSize) { "Encrypted health cache is invalid" }
        val iv = packet.copyOfRange(1, 1 + ivSize)
        val encrypted = packet.copyOfRange(1 + ivSize, packet.size)
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.DECRYPT_MODE, secretKey(), GCMParameterSpec(128, iv))
        return cipher.doFinal(encrypted)
    }
}

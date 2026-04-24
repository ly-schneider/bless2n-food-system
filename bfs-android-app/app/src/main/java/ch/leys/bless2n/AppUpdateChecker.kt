package ch.leys.bless2n

import android.app.Activity
import android.app.AlertDialog
import android.content.Intent
import android.util.Log
import android.widget.Toast
import androidx.core.content.FileProvider
import org.json.JSONObject
import java.io.File
import java.io.FileOutputStream
import java.net.HttpURLConnection
import java.net.URL
import java.security.MessageDigest

private const val TAG = "AppUpdateChecker"

class AppUpdateChecker(private val activity: Activity) {

    fun checkForUpdate() {
        val url = BuildConfig.UPDATE_CHECK_URL
        if (url.isBlank()) {
            Log.w(TAG, "UPDATE_CHECK_URL is blank; skipping update check")
            return
        }

        Log.i(TAG, "Checking for update: url=$url currentVersionCode=${BuildConfig.VERSION_CODE}")

        Thread {
            try {
                val conn = URL(url).openConnection() as HttpURLConnection
                conn.connectTimeout = 10_000
                conn.readTimeout = 10_000
                conn.requestMethod = "GET"
                conn.setRequestProperty("Accept", "application/json")

                val code = conn.responseCode
                if (code != 200) {
                    Log.w(TAG, "Update endpoint returned HTTP $code")
                    conn.disconnect()
                    return@Thread
                }

                val body = conn.inputStream.bufferedReader().use { it.readText() }
                conn.disconnect()

                val json = JSONObject(body)
                val remoteVersionCode = json.optInt("versionCode", 0)
                val remoteVersionName = json.optString("versionName", "")
                val downloadUrl = json.optString("downloadUrl", "")
                val sha256 = json.optString("sha256", "")

                Log.i(
                    TAG,
                    "Remote release: versionCode=$remoteVersionCode versionName=$remoteVersionName downloadUrl=$downloadUrl"
                )

                if (remoteVersionCode <= BuildConfig.VERSION_CODE) {
                    Log.i(
                        TAG,
                        "Remote versionCode $remoteVersionCode <= current ${BuildConfig.VERSION_CODE}; no update"
                    )
                    return@Thread
                }
                if (downloadUrl.isBlank()) {
                    Log.w(TAG, "Remote release missing downloadUrl; skipping")
                    return@Thread
                }

                activity.runOnUiThread {
                    AlertDialog.Builder(activity)
                        .setTitle("Update verfügbar")
                        .setMessage("Version $remoteVersionName ist verfügbar. Jetzt installieren?")
                        .setPositiveButton("Installieren") { _, _ ->
                            downloadAndInstall(downloadUrl, sha256)
                        }
                        .setNegativeButton("Später", null)
                        .setCancelable(true)
                        .show()
                }
            } catch (t: Throwable) {
                Log.e(TAG, "Update check failed", t)
            }
        }.start()
    }

    private fun downloadAndInstall(downloadUrl: String, expectedSha256: String) {
        Thread {
            try {
                val conn = URL(downloadUrl).openConnection() as HttpURLConnection
                conn.connectTimeout = 30_000
                conn.readTimeout = 60_000
                conn.instanceFollowRedirects = true

                if (conn.responseCode != 200) {
                    conn.disconnect()
                    showToast("Download fehlgeschlagen")
                    return@Thread
                }

                val updateDir = File(activity.cacheDir, "updates")
                updateDir.mkdirs()
                val apkFile = File(updateDir, "update.apk")

                conn.inputStream.use { input ->
                    FileOutputStream(apkFile).use { output ->
                        input.copyTo(output)
                    }
                }
                conn.disconnect()

                if (expectedSha256.isNotBlank()) {
                    val digest = MessageDigest.getInstance("SHA-256")
                    apkFile.inputStream().use { input ->
                        val buffer = ByteArray(8192)
                        var read: Int
                        while (input.read(buffer).also { read = it } != -1) {
                            digest.update(buffer, 0, read)
                        }
                    }
                    val actualSha256 = digest.digest().joinToString("") { "%02x".format(it) }
                    if (!actualSha256.equals(expectedSha256, ignoreCase = true)) {
                        apkFile.delete()
                        showToast("Update-Prüfung fehlgeschlagen")
                        return@Thread
                    }
                }

                val uri = FileProvider.getUriForFile(
                    activity,
                    "${activity.packageName}.fileprovider",
                    apkFile
                )

                val intent = Intent(Intent.ACTION_VIEW).apply {
                    setDataAndType(uri, "application/vnd.android.package-archive")
                    addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                    addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                }
                activity.startActivity(intent)
            } catch (_: Throwable) {
                showToast("Update fehlgeschlagen")
            }
        }.start()
    }

    private fun showToast(message: String) {
        activity.runOnUiThread {
            Toast.makeText(activity, message, Toast.LENGTH_SHORT).show()
        }
    }
}

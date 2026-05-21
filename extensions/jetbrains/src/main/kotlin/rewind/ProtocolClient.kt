package com.rewind

import com.google.gson.Gson
import com.google.gson.annotations.SerializedName
import java.net.HttpURLConnection
import java.net.URI
import java.net.URLEncoder
import java.nio.charset.StandardCharsets

data class IDEEventData(
    val type: String,
    val timestamp: String,
    val file: String? = null,
    val language: String? = null,
    @SerializedName("lines_added") val linesAdded: Int? = null,
    @SerializedName("lines_removed") val linesRemoved: Int? = null,
    @SerializedName("exit_code") val exitCode: Int? = null,
    val message: String? = null,
    val metadata: Map<String, Any>? = null,
    @SerializedName("content_snapshot") val contentSnapshot: String? = null,
    @SerializedName("session_id") val sessionId: String? = null
)

data class IDEEvent(
    val protocol: String = "rewind-ide-v1",
    val ide: String,
    val version: String,
    val project: String,
    @SerializedName("project_path") val projectPath: String,
    val event: IDEEventData
)

class ProtocolClient(private val port: Int = 9876, private val ideName: String, private val ideVersion: String) {
    private val gson = Gson()
    private val baseUrl = "http://localhost:$port"
    private var enabled = false // Opt-in: only send if enabled

    fun setEnabled(on: Boolean) { enabled = on }

    fun sendEvent(type: String, projectName: String, projectPath: String, extra: Map<String, Any?> = emptyMap()) {
        if (!enabled) return
        try {
            val event = IDEEvent(
                ide = ideName,
                version = ideVersion,
                project = projectName,
                projectPath = projectPath,
                event = IDEEventData(
                    type = type,
                    timestamp = java.time.Instant.now().toString(),
                    file = extra["file"] as? String,
                    language = extra["language"] as? String,
                    linesAdded = extra["lines_added"] as? Int,
                    linesRemoved = extra["lines_removed"] as? Int,
                    exitCode = extra["exit_code"] as? Int,
                    message = extra["message"] as? String,
                    metadata = extra["metadata"] as? Map<String, Any>,
                    contentSnapshot = extra["content_snapshot"] as? String,
                    sessionId = extra["session_id"] as? String
                )
            )
            postJson("/", gson.toJson(event))
        } catch (e: Exception) {
            // Server unreachable — silently ignore
        }
    }

    fun checkHealth(): Boolean {
        return try {
            val resp = get("/health")
            resp.contains("\"healthy\"")
        } catch (e: Exception) { false }
    }

    fun getStatus(): String? {
        return try { get("/status") } catch (e: Exception) { null }
    }

    private fun postJson(path: String, json: String) {
        val url = URI(baseUrl + path).toURL()
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "POST"
        conn.doOutput = true
        conn.setRequestProperty("Content-Type", "application/json")
        conn.connectTimeout = 3000
        conn.readTimeout = 3000
        conn.outputStream.use { os -> os.write(json.toByteArray(StandardCharsets.UTF_8)) }
        conn.responseCode // trigger request
        conn.disconnect()
    }

    private fun get(path: String): String {
        val url = URI(baseUrl + path).toURL()
        val conn = url.openConnection() as HttpURLConnection
        conn.requestMethod = "GET"
        conn.connectTimeout = 3000
        conn.readTimeout = 3000
        return conn.inputStream.bufferedReader().readText().also { conn.disconnect() }
    }
}
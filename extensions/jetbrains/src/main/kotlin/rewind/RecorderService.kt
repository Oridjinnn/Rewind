package com.rewind

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.components.Service
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile

/**
 * Application-level service that hooks into IDE events and sends them to Rewind server.
 */
@Service(Service.Level.APP)
class RecorderService {
    private val client: ProtocolClient
    private var recordingEnabled = false

    init {
        val ideName = detectIDEName()
        val app = ApplicationManager.getApplication()
        client = ProtocolClient(9876, ideName, app.getPlatformVersion())
    }

    fun enable() { client.setEnabled(true); recordingEnabled = true }
    fun disable() { client.setEnabled(false); recordingEnabled = false }
    fun isEnabled() = recordingEnabled

    fun recordFileOpen(file: VirtualFile, project: Project) {
        client.sendEvent("file_open", projectName(project), projectPath(project), mapOf(
            "file" to file.path,
            "language" to file.fileType.name.lowercase()
        ))
    }

    fun recordFileSave(file: VirtualFile, project: Project) {
        client.sendEvent("file_save", projectName(project), projectPath(project), mapOf(
            "file" to file.path,
            "language" to file.fileType.name.lowercase()
        ))
    }

    fun recordFileClose(file: VirtualFile, project: Project) {
        client.sendEvent("file_close", projectName(project), projectPath(project), mapOf(
            "file" to file.path,
            "language" to file.fileType.name.lowercase()
        ))
    }

    fun recordBuildStart(project: Project, name: String) {
        client.sendEvent("build_start", projectName(project), projectPath(project), mapOf(
            "message" to name
        ))
    }

    fun recordBuildEnd(project: Project, name: String, exitCode: Int) {
        val type = if (exitCode == 0) "build_end" else "build_error"
        client.sendEvent(type, projectName(project), projectPath(project), mapOf(
            "message" to name,
            "exit_code" to exitCode
        ))
    }

    fun recordTestRun(project: Project, name: String) {
        client.sendEvent("test_run", projectName(project), projectPath(project), mapOf("message" to name))
    }

    fun recordTestResult(project: Project, name: String, passed: Boolean) {
        val type = if (passed) "test_pass" else "test_fail"
        client.sendEvent(type, projectName(project), projectPath(project), mapOf("message" to name))
    }

    fun recordGitOp(project: Project, op: String, branch: String) {
        client.sendEvent(op, projectName(project), projectPath(project), mapOf(
            "message" to "Git $op: $branch",
            "metadata" to mapOf("branch" to branch)
        ))
    }

    fun recordTerminalCmd(project: Project, cmd: String, exitCode: Int) {
        client.sendEvent("terminal_cmd", projectName(project), projectPath(project), mapOf(
            "message" to cmd,
            "exit_code" to exitCode
        ))
    }

    private fun projectName(project: Project) = project.name
    private fun projectPath(project: Project) = project.basePath ?: ""

    private fun detectIDEName(): String {
        val name = ApplicationManager.getApplication().getName()
        return when {
            name.contains("GoLand", true) -> "goland"
            name.contains("PyCharm", true) -> "pycharm"
            name.contains("WebStorm", true) -> "webstorm"
            name.contains("Android", true) -> "android-studio"
            name.contains("IDEA", true) || name.contains("IntelliJ", true) -> "intellij-idea"
            else -> "intellij-idea"
        }
    }
}
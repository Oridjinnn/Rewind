package com.rewind

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project
import com.intellij.openapi.project.ProjectManagerListener
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.openapi.fileEditor.FileEditorManagerListener
import com.intellij.openapi.fileEditor.FileEditorManager
import com.intellij.openapi.startup.StartupActivity
import com.intellij.openapi.project.ProjectManager
import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType

/**
 * Plugin entry point — registers all event listeners.
 */
class RewindPlugin : StartupActivity {
    override fun runActivity(project: Project) {
        val recorder = ApplicationManager.getApplication().service<RecorderService>()

        // File open/close listeners
        project.messageBus.connect().subscribe(
            FileEditorManagerListener.FILE_EDITOR_MANAGER,
            object : FileEditorManagerListener {
                override fun fileOpened(source: FileEditorManager, file: VirtualFile) {
                    recorder.recordFileOpen(file, source.project)
                }
                override fun fileClosed(source: FileEditorManager, file: VirtualFile) {
                    recorder.recordFileClose(file, source.project)
                }
            }
        )

        // Notify that Rewind is active
        val notif = NotificationGroupManager.getInstance()
            .getNotificationGroup("Rewind Notifications")
        notif.createNotification(
            "Rewind IDE active—recording activity to local SQLite database.\n" +
            "Opt-in required: rewind ide permissions ${recorder.detectIDEName()} on",
            NotificationType.INFORMATION
        ).notify(project)
    }
}

// Action: Toggle recording
class ToggleRecordingAction : com.intellij.openapi.actionSystem.AnAction() {
    override fun actionPerformed(e: com.intellij.openapi.actionSystem.AnActionEvent) {
        val recorder = ApplicationManager.getApplication().service<RecorderService>()
        if (recorder.isEnabled()) {
            recorder.disable()
        } else {
            recorder.enable()
        }
        val notif = NotificationGroupManager.getInstance()
            .getNotificationGroup("Rewind Notifications")
        notif.createNotification(
            "Rewind recording ${if (recorder.isEnabled()) "enabled" else "paused"}",
            NotificationType.INFORMATION
        ).notify(e.project)
    }
}

// Action: Show status
class ShowStatusAction : com.intellij.openapi.actionSystem.AnAction() {
    override fun actionPerformed(e: com.intellij.openapi.actionSystem.AnActionEvent) {
        val recorder = ApplicationManager.getApplication().service<RecorderService>()
        val msg = "Rewind: ${if (recorder.isEnabled()) "Recording" else "Paused"}"
        val notif = NotificationGroupManager.getInstance()
            .getNotificationGroup("Rewind Notifications")
        notif.createNotification(msg, NotificationType.INFORMATION).notify(e.project)
    }
}

// Action: Search history
class SearchHistoryAction : com.intellij.openapi.actionSystem.AnAction() {
    override fun actionPerformed(e: com.intellij.openapi.actionSystem.AnActionEvent) {
        // Opens terminal with rewind search
        val project = e.project ?: return
        val terminal = com.intellij.openapi.wm.ToolWindowManager.getInstance(project)
            .getToolWindow("Terminal") ?: return
        terminal.show()
        // User must have rewind CLI in PATH
    }
}

// Status bar widget factory
class RewindStatusBarFactory : com.intellij.openapi.wm.StatusBarWidgetFactory {
    override fun getId() = "RewindStatus"
    override fun getDisplayName() = "Rewind Status"
    override fun isAvailable(project: Project) = true

    override fun createWidget(project: Project) =
        com.intellij.openapi.wm.impl.status.widget.StatusBarWidgetWrapper(
            com.intellij.openapi.wm.StatusBarWidget.ID { "RewindStatus" },
            object : com.intellij.openapi.wm.StatusBarWidget, com.intellij.openapi.wm.CustomStatusBarWidget {
                private var enabled = false
                override fun ID() = com.intellij.openapi.wm.StatusBarWidget.ID { "RewindStatus" }
                override fun getPresentation() = object : com.intellij.openapi.wm.StatusBarWidget.WidgetPresentation {
                    override fun getTooltipText() = "Rewind IDE Recording"
                    override fun getPresentationType() = com.intellij.openapi.wm.StatusBarWidget.PresentationType.TEXT
                    override fun getText(): String {
                        val recorder = try { ApplicationManager.getApplication().service<RecorderService>() } catch (_: Exception) { null }
                        return if (recorder?.isEnabled() == true) "⚫ Rewind" else "⚪ Rewind"
                    }
                }
                override fun install(frame: com.intellij.openapi.wm.StatusBar) { enabled = true }
                override fun dispose() { enabled = false }
            }
        )

    override fun disposeWidget(widget: com.intellij.openapi.wm.StatusBarWidget) {}
}

// Configurable settings panel
class RewindConfigurable : com.intellij.openapi.options.Configurable {
    private var enabled = false
    override fun getDisplayName() = "Rewind IDE"
    override fun createComponent() = com.intellij.ui.components.JBCheckBox("Enable activity recording", enabled).apply {
        addActionListener { enabled = isSelected }
    }
    override fun isModified() = true
    override fun apply() {
        val recorder = ApplicationManager.getApplication().service<RecorderService>()
        if (enabled) recorder.enable() else recorder.disable()
    }
}
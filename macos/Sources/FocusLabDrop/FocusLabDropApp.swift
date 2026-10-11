import AppKit
import SwiftUI

@main
struct FocusLabDropApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var delegate
    private let model = AppModel.shared

    var body: some Scene {
        MenuBarExtra {
            MenuContent(model: model, settings: model.settings, server: model.server)
        } label: {
            MenuBarLabel(server: model.server)
        }
    }
}

private struct MenuBarLabel: View {
    @ObservedObject var server: ServerController
    var body: some View { Image(nsImage: StatusImage.dot(server.state)) }
}

@MainActor
final class AppDelegate: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        NSApp.setActivationPolicy(.accessory) // menu bar only, no Dock icon
        let model = AppModel.shared
        if model.settings.startServerOnLaunch { model.startServer() }
        if !startedAsLoginItem() { model.showWindow() }
    }

    func applicationShouldHandleReopen(_ sender: NSApplication, hasVisibleWindows flag: Bool) -> Bool {
        AppModel.shared.showWindow()
        return true
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool { false }

    func applicationWillTerminate(_ notification: Notification) { AppModel.shared.server.stop() }

    /// At login the app should stay quiet in the menu bar. A person opening it gets the window.
    private func startedAsLoginItem() -> Bool {
        if let event = NSAppleEventManager.shared().currentAppleEvent,
           event.paramDescriptor(forKeyword: AEKeyword(keyAEPropData))?.enumCodeValue == OSType(keyAELaunchedAsLogInItem) {
            return true
        }
        return ProcessInfo.processInfo.systemUptime < 240 // just booted: nobody has opened it by hand yet
    }
}

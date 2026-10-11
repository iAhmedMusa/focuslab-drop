import AppKit
import FocusLabCore
import ServiceManagement
import SwiftUI

@MainActor
final class AppModel {
    static let shared = AppModel()
    let settings = AppSettings()
    let server = ServerController()
    private var mainWindow: NSWindow?
    private var posterWindow: NSWindow?

    // MARK: addresses and links

    var addresses: [LANAddress] { lanAddresses() }
    var selectedAddress: LANAddress? {
        let all = addresses
        return all.first { $0.ip == settings.selectedIP } ?? all.first
    }
    var customerLink: String? { selectedAddress.map { customerURL(ip: $0.ip, port: settings.port) } }

    // MARK: server

    func startServer() { server.start(dir: settings.folderPath, port: settings.port) }

    func setServer(on: Bool) {
        if on { startServer() } else { server.stop() }
    }

    func restartIfOn() {
        guard server.isOn else { return }
        server.stop()
        startServer()
    }

    // MARK: actions

    func chooseFolder() {
        let panel = NSOpenPanel()
        panel.canChooseDirectories = true
        panel.canChooseFiles = false
        panel.canCreateDirectories = true
        panel.prompt = "Use this folder"
        panel.message = "Choose where customer photos are saved."
        panel.directoryURL = URL(fileURLWithPath: settings.folderPath)
        guard panel.runModal() == .OK, let url = panel.url else { return }
        settings.folderPath = url.path
        restartIfOn()
    }

    func applyPort(_ text: String) -> Bool {
        guard let p = Int(text.trimmingCharacters(in: .whitespaces)), isValidPort(p) else {
            alert("Enter a port between 1024 and 65535.")
            return false
        }
        guard p != settings.port else { return true }
        settings.port = p
        restartIfOn()
        return true
    }

    func openCounter() { NSWorkspace.shared.open(URL(string: counterURL(port: settings.port))!) }

    func openFolder() {
        try? FileManager.default.createDirectory(atPath: settings.folderPath, withIntermediateDirectories: true)
        NSWorkspace.shared.open(URL(fileURLWithPath: settings.folderPath))
    }

    func openLog() { NSWorkspace.shared.open(ServerController.logURL) }

    func copyLink() {
        guard let link = customerLink else { return }
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(link, forType: .string)
    }

    func setLaunchAtLogin(_ on: Bool) {
        if let error = settings.setLaunchAtLogin(on) {
            alert("macOS did not allow this change.\n\(error)")
        } else if on, settings.loginStatus == .requiresApproval {
            let a = NSAlert()
            a.messageText = "One more step"
            a.informativeText = "Allow FocusLab Drop in System Settings, General, Login Items."
            a.addButton(withTitle: "Open Login Items")
            a.addButton(withTitle: "Later")
            if a.runModal() == .alertFirstButtonReturn { SMAppService.openSystemSettingsLoginItems() }
        }
    }

    func alert(_ message: String) {
        NSApp.activate(ignoringOtherApps: true)
        let a = NSAlert()
        a.messageText = message
        a.runModal()
    }

    func confirmQuit() {
        if server.isOn {
            NSApp.activate(ignoringOtherApps: true)
            let a = NSAlert()
            a.messageText = "Quit FocusLab Drop?"
            a.informativeText = "The server will stop, and customers cannot send photos until the app is opened again."
            a.addButton(withTitle: "Quit")
            a.addButton(withTitle: "Cancel")
            guard a.runModal() == .alertFirstButtonReturn else { return }
        }
        NSApp.terminate(nil)
    }

    // MARK: windows

    func showWindow() {
        if mainWindow == nil {
            let w = NSWindow(contentRect: .zero, styleMask: [.titled, .closable, .miniaturizable], backing: .buffered, defer: false)
            w.title = "FocusLab Drop"
            w.isReleasedWhenClosed = false
            w.contentViewController = NSHostingController(rootView: MainWindow(model: self, settings: settings, server: server))
            w.center()
            mainWindow = w
        }
        settings.refreshLoginStatus()
        NSApp.activate(ignoringOtherApps: true)
        mainWindow?.makeKeyAndOrderFront(nil)
    }

    func showPoster() {
        guard let link = customerLink else { return alert("No network address found. Connect this Mac to the shop Wi-Fi first.") }
        posterWindow?.close()
        let w = NSWindow(contentRect: .zero, styleMask: [.titled, .closable], backing: .buffered, defer: false)
        w.title = "Upload QR poster"
        w.isReleasedWhenClosed = false
        w.contentViewController = NSHostingController(rootView: PosterPreview(url: link))
        w.center()
        posterWindow = w
        NSApp.activate(ignoringOtherApps: true)
        w.makeKeyAndOrderFront(nil)
    }
}

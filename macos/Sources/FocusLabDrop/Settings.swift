import AppKit
import FocusLabCore
import ServiceManagement

@MainActor
final class AppSettings: ObservableObject {
    private let defaults = UserDefaults.standard

    @Published var folderPath: String { didSet { defaults.set(folderPath, forKey: "folderPath") } }
    @Published var port: Int { didSet { defaults.set(port, forKey: "port") } }
    @Published var startServerOnLaunch: Bool { didSet { defaults.set(startServerOnLaunch, forKey: "startServerOnLaunch") } }
    @Published var selectedIP: String { didSet { defaults.set(selectedIP, forKey: "selectedIP") } }
    @Published private(set) var loginStatus: SMAppService.Status = SMAppService.mainApp.status

    init() {
        defaults.register(defaults: [
            "folderPath": NSHomeDirectory() + "/Prints",
            "port": 6789,
            "startServerOnLaunch": true,
            "selectedIP": "",
        ])
        folderPath = defaults.string(forKey: "folderPath") ?? NSHomeDirectory() + "/Prints"
        port = defaults.integer(forKey: "port")
        startServerOnLaunch = defaults.bool(forKey: "startServerOnLaunch")
        selectedIP = defaults.string(forKey: "selectedIP") ?? ""
    }

    var launchAtLogin: Bool { loginStatus == .enabled }

    /// Registers or removes the app as a login item. Returns an error message if macOS refuses.
    func setLaunchAtLogin(_ on: Bool) -> String? {
        defer { loginStatus = SMAppService.mainApp.status }
        do {
            if on { try SMAppService.mainApp.register() } else { try SMAppService.mainApp.unregister() }
            return nil
        } catch {
            return error.localizedDescription
        }
    }

    func refreshLoginStatus() { loginStatus = SMAppService.mainApp.status }
}

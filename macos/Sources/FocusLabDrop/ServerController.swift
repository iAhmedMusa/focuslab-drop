import AppKit
import FocusLabCore

/// Runs the bundled Go server as a child process and reports whether it is really answering.
@MainActor
final class ServerController: ObservableObject {
    enum State: Equatable {
        case stopped, starting, running
        case failed(String)
    }

    @Published private(set) var state: State = .stopped

    private var process: Process?
    private var stdinPipe: Pipe?   // held open on purpose: when the app dies the pipe closes and the server exits
    private var runID = 0
    private var lastLine = ""

    static let logURL = FileManager.default.homeDirectoryForCurrentUser
        .appendingPathComponent("Library/Logs/FocusLab Drop/server.log")

    var isOn: Bool { state == .starting || state == .running }

    private var serverBinary: URL? {
        if let env = ProcessInfo.processInfo.environment["FOCUSLAB_SERVER_BIN"] { return URL(fileURLWithPath: env) }
        return Bundle.main.resourceURL?.appendingPathComponent("focuslab-drop")
    }

    func start(dir: String, port: Int) {
        guard !isOn else { return }
        guard let bin = serverBinary, FileManager.default.isExecutableFile(atPath: bin.path) else {
            return fail("The server program is missing from the app. Reinstall FocusLab Drop.")
        }
        guard isValidPort(port) else { return fail("Port \(port) is not valid. Use 1024 to 65535.") }
        guard !isPortInUse(port) else {
            return fail("Port \(port) is already in use. Another copy may be running. Quit it, or change the port under Advanced.")
        }

        let p = Process()
        p.executableURL = bin
        p.arguments = serverArguments(dir: dir, port: port)
        let input = Pipe(), output = Pipe()
        p.standardInput = input
        p.standardOutput = output
        p.standardError = output
        output.fileHandleForReading.readabilityHandler = { [weak self] h in
            let data = h.availableData
            if data.isEmpty { h.readabilityHandler = nil; return }
            if let text = String(data: data, encoding: .utf8) { Task { @MainActor in self?.log(text) } }
        }
        runID += 1
        let id = runID
        p.terminationHandler = { [weak self] proc in
            Task { @MainActor in self?.didExit(id: id, code: proc.terminationStatus) }
        }

        resetLog()
        lastLine = ""
        do { try p.run() } catch { return fail("Could not start the server: \(error.localizedDescription)") }
        process = p
        stdinPipe = input
        state = .starting
        waitUntilAnswering(id: id, port: port)
    }

    func stop() {
        guard let p = process else { state = .stopped; return }
        runID += 1 // the exit handler of this run is now stale and must not report a failure
        p.terminate()
        let deadline = Date().addingTimeInterval(2)
        while p.isRunning && Date() < deadline { usleep(20_000) }
        if p.isRunning { kill(p.processIdentifier, SIGKILL) }
        process = nil
        stdinPipe = nil
        state = .stopped
    }

    // MARK: - private

    private func fail(_ message: String) { state = .failed(message) }

    private func didExit(id: Int, code: Int32) {
        guard id == runID else { return }
        process = nil
        stdinPipe = nil
        let detail = lastLine.isEmpty ? "exit code \(code)" : lastLine
        fail("The server stopped by itself (\(detail)). Check the folder and try again.")
    }

    private func waitUntilAnswering(id: Int, port: Int) {
        Task { @MainActor in
            for _ in 0..<40 {
                guard id == runID, process != nil else { return }
                if await Self.answers(port: port) {
                    if id == runID { state = .running }
                    return
                }
                try? await Task.sleep(nanoseconds: 250_000_000)
            }
            guard id == runID else { return }
            stop()
            fail("The server did not start in time. Try again.")
        }
    }

    private nonisolated static func answers(port: Int) async -> Bool {
        var req = URLRequest(url: URL(string: "http://127.0.0.1:\(port)/drop")!)
        req.timeoutInterval = 1
        guard let (_, resp) = try? await URLSession.shared.data(for: req) else { return false }
        return (resp as? HTTPURLResponse)?.statusCode == 200
    }

    private func resetLog() {
        let dir = Self.logURL.deletingLastPathComponent()
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        try? Data().write(to: Self.logURL)
    }

    private func log(_ text: String) {
        if let line = text.split(separator: "\n").last { lastLine = String(line).trimmingCharacters(in: .whitespaces) }
        guard let h = try? FileHandle(forWritingTo: Self.logURL) else { return }
        defer { try? h.close() }
        _ = try? h.seekToEnd()
        try? h.write(contentsOf: Data(text.utf8))
    }
}

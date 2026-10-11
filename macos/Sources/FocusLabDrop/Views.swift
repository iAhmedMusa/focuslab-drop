import AppKit
import FocusLabCore
import SwiftUI

struct MainWindow: View {
    let model: AppModel
    @ObservedObject var settings: AppSettings
    @ObservedObject var server: ServerController
    @State private var portText = ""
    @State private var refreshTick = 0
    @State private var pulse = false
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    private let refresh = Timer.publish(every: 5, on: .main, in: .common).autoconnect()

    var body: some View {
        let addresses = model.addresses
        let _ = refreshTick // re-read the network addresses every few seconds
        VStack(alignment: .leading, spacing: 20) {
            header
            statusCard
            folderSection
            linkSection(addresses)
            optionsSection
            footer
        }
        .padding(24)
        .frame(width: 460)
        .onAppear { portText = String(settings.port) }
        .onReceive(refresh) { _ in refreshTick += 1 }
    }

    // MARK: sections

    private var header: some View {
        HStack(spacing: 12) {
            if let logo = Brand.logo { Image(nsImage: logo).resizable().frame(width: 44, height: 44) }
            VStack(alignment: .leading, spacing: 2) {
                Text("Focus Digital Color Lab").font(.system(size: 17, weight: .semibold))
                Text("Photo drop-off").font(.callout).foregroundStyle(.secondary)
            }
        }
    }

    private var statusCard: some View {
        HStack(spacing: 14) {
            Circle()
                .fill(server.state.color)
                .frame(width: 18, height: 18)
                .shadow(color: server.state == .running ? server.state.color.opacity(0.7) : .clear, radius: 7)
                .opacity(server.state == .starting && pulse ? 0.35 : 1)
                .animation(server.state == .starting && !reduceMotion ? .easeInOut(duration: 0.7).repeatForever() : .default, value: pulse)
                .onChange(of: server.state) { state in pulse = (state == .starting) }
            VStack(alignment: .leading, spacing: 3) {
                Text(server.state.title).font(.title3.weight(.semibold))
                Text(server.state.detail)
                    .font(.callout)
                    .foregroundStyle(isFailed ? Color.red : Color.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
            Spacer(minLength: 8)
            Toggle("Server", isOn: Binding(get: { server.isOn }, set: { model.setServer(on: $0) }))
                .toggleStyle(.switch)
                .tint(Brand.gold)
                .labelsHidden()
                .disabled(server.state == .starting)
        }
        .padding(16)
        .background(RoundedRectangle(cornerRadius: Brand.radius).fill(Color(nsColor: .controlBackgroundColor)))
        .overlay(RoundedRectangle(cornerRadius: Brand.radius).stroke(Color(nsColor: .separatorColor)))
    }

    private var isFailed: Bool { if case .failed = server.state { return true } else { return false } }

    private var folderSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Save photos to").font(.callout.weight(.semibold))
            HStack {
                Text(settings.folderPath)
                    .lineLimit(1).truncationMode(.middle)
                    .padding(.horizontal, 10).frame(maxWidth: .infinity, minHeight: 32, alignment: .leading)
                    .background(RoundedRectangle(cornerRadius: 8).fill(Color(nsColor: .controlBackgroundColor)))
                    .overlay(RoundedRectangle(cornerRadius: 8).stroke(Color(nsColor: .separatorColor)))
                Button("Choose...") { model.chooseFolder() }
                Button("Show") { model.openFolder() }
            }
        }
    }

    private func linkSection(_ addresses: [LANAddress]) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Customer link").font(.callout.weight(.semibold))
            if addresses.isEmpty {
                Text("No network address found. Connect this Mac to the shop Wi-Fi.")
                    .font(.callout).foregroundStyle(Color.orange)
            } else {
                if addresses.count > 1 {
                    Picker("Network", selection: Binding(get: { model.selectedAddress?.ip ?? "" }, set: { settings.selectedIP = $0 })) {
                        ForEach(addresses) { Text("\($0.ip)  (\($0.interface))").tag($0.ip) }
                    }
                }
                HStack {
                    Text(model.customerLink ?? "")
                        .font(.system(.callout, design: .monospaced)).textSelection(.enabled)
                        .lineLimit(1).truncationMode(.middle)
                        .padding(.horizontal, 10).frame(maxWidth: .infinity, minHeight: 32, alignment: .leading)
                        .background(RoundedRectangle(cornerRadius: 8).fill(Color(nsColor: .controlBackgroundColor)))
                        .overlay(RoundedRectangle(cornerRadius: 8).stroke(Color(nsColor: .separatorColor)))
                    Button("Copy") { model.copyLink() }
                }
            }
            HStack {
                Button("Print QR poster...") { model.showPoster() }
                    .buttonStyle(GoldButtonStyle()).disabled(addresses.isEmpty)
                Button("Open counter page") { model.openCounter() }.disabled(server.state != .running)
            }
            .padding(.top, 4)
        }
    }

    private var optionsSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            Toggle("Launch at login", isOn: Binding(get: { settings.launchAtLogin }, set: { model.setLaunchAtLogin($0) }))
                .tint(Brand.gold)
            if settings.loginStatus == .requiresApproval {
                Text("Waiting for your OK in System Settings, General, Login Items.")
                    .font(.caption).foregroundStyle(Color.orange)
            }
            Toggle("Start the server when this app opens", isOn: $settings.startServerOnLaunch)
                .tint(Brand.gold)
            DisclosureGroup("Advanced") {
                HStack {
                    Text("Port")
                    TextField("6789", text: $portText).frame(width: 80).onSubmit { applyPort() }
                    Button("Apply") { applyPort() }
                }
                .padding(.top, 6)
                Text("Changing the port makes printed QR codes stop working.")
                    .font(.caption).foregroundStyle(.secondary)
            }
        }
    }

    private var footer: some View {
        HStack {
            Button("Show server log") { model.openLog() }.buttonStyle(.link)
            Spacer()
            Button("Quit") { model.confirmQuit() }.buttonStyle(.link)
        }
        .font(.caption)
    }

    private func applyPort() {
        if !model.applyPort(portText) { portText = String(settings.port) }
    }
}

struct MenuContent: View {
    let model: AppModel
    @ObservedObject var settings: AppSettings
    @ObservedObject var server: ServerController

    var body: some View {
        Label { Text(server.state.title) } icon: { Image(nsImage: StatusImage.dot(server.state)) }
        Button(server.isOn ? "Turn server off" : "Turn server on") { model.setServer(on: !server.isOn) }
            .disabled(server.state == .starting)
        Divider()
        Button("Open FocusLab Drop...") { model.showWindow() }
        Button("Open counter page") { model.openCounter() }.disabled(server.state != .running)
        Button("Open photos folder") { model.openFolder() }
        Button("Copy customer link") { model.copyLink() }.disabled(model.customerLink == nil)
        Button("Print QR poster...") { model.showPoster() }.disabled(model.customerLink == nil)
        Divider()
        Toggle("Launch at login", isOn: Binding(get: { settings.launchAtLogin }, set: { model.setLaunchAtLogin($0) }))
        Divider()
        Button("Quit FocusLab Drop") { model.confirmQuit() }.keyboardShortcut("q")
    }
}

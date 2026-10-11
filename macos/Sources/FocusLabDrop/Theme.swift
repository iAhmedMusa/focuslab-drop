import AppKit
import SwiftUI

enum Brand {
    /// Sky blue, the same accent as the web pages (dark theme values).
    static let accent = Color(red: 124 / 255, green: 195 / 255, blue: 236 / 255)
    static let accentInk = Color(red: 7 / 255, green: 32 / 255, blue: 58 / 255)
    static let on = Color(red: 0.18, green: 0.80, blue: 0.44)
    static let radius: CGFloat = 12

    static var logo: NSImage? {
        Bundle.main.url(forResource: "logo", withExtension: "png").flatMap { NSImage(contentsOf: $0) }
    }
}

extension ServerController.State {
    var title: String {
        switch self {
        case .running: return "Server is on"
        case .stopped: return "Server is off"
        case .starting: return "Starting..."
        case .failed: return "Server problem"
        }
    }

    var detail: String {
        switch self {
        case .running: return "Customers can send photos."
        case .stopped: return "Turn it on to receive photos."
        case .starting: return "Getting ready."
        case .failed(let message): return message
        }
    }

    var nsColor: NSColor {
        switch self {
        case .running: return .systemGreen
        case .starting: return NSColor(calibratedRed: 0.96, green: 0.66, blue: 0.05, alpha: 1)
        case .stopped, .failed: return .systemRed
        }
    }

    var color: Color { Color(nsColor: nsColor) }
}

enum StatusImage {
    /// A coloured dot for the menu bar. It is not a template image, so macOS keeps the colour.
    static func dot(_ state: ServerController.State) -> NSImage {
        let image = NSImage(size: NSSize(width: 18, height: 18), flipped: false) { rect in
            let circle = NSBezierPath(ovalIn: rect.insetBy(dx: 2.5, dy: 2.5))
            state.nsColor.setFill()
            circle.fill()
            NSColor.black.withAlphaComponent(0.25).setStroke()
            circle.lineWidth = 1
            circle.stroke()
            if case .failed = state {
                let mark = NSAttributedString(string: "!", attributes: [
                    .font: NSFont.systemFont(ofSize: 11, weight: .heavy), .foregroundColor: NSColor.white,
                ])
                let s = mark.size()
                mark.draw(at: NSPoint(x: rect.midX - s.width / 2, y: rect.midY - s.height / 2))
            }
            return true
        }
        image.isTemplate = false
        return image
    }
}

struct AccentButtonStyle: ButtonStyle {
    @Environment(\.isEnabled) private var enabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.system(size: 13, weight: .semibold))
            .foregroundStyle(Brand.accentInk)
            .padding(.horizontal, 16)
            .frame(minHeight: 34)
            .background(RoundedRectangle(cornerRadius: Brand.radius).fill(Brand.accent))
            .opacity(enabled ? (configuration.isPressed ? 0.8 : 1) : 0.4)
            .scaleEffect(configuration.isPressed ? 0.98 : 1)
    }
}

/// Round power button: green when the server is on, gray when off.
struct PowerButtonStyle: ButtonStyle {
    let isOn: Bool
    @Environment(\.isEnabled) private var enabled

    func makeBody(configuration: Configuration) -> some View {
        Image(systemName: "power")
            .font(.system(size: 19, weight: .bold))
            .foregroundStyle(isOn ? Color.white : Color.secondary)
            .frame(width: 46, height: 46)
            .background(Circle().fill(isOn ? Brand.on : Color(nsColor: .quaternaryLabelColor)))
            .shadow(color: isOn ? Brand.on.opacity(0.55) : .clear, radius: 8)
            .opacity(enabled ? (configuration.isPressed ? 0.8 : 1) : 0.4)
            .scaleEffect(configuration.isPressed ? 0.96 : 1)
    }
}

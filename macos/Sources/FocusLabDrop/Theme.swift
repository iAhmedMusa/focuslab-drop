import AppKit
import SwiftUI

enum Brand {
    static let navy = Color(red: 13 / 255, green: 27 / 255, blue: 42 / 255)
    static let gold = Color(red: 212 / 255, green: 160 / 255, blue: 23 / 255)
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

struct GoldButtonStyle: ButtonStyle {
    @Environment(\.isEnabled) private var enabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.system(size: 13, weight: .semibold))
            .foregroundStyle(Brand.navy)
            .padding(.horizontal, 16)
            .frame(minHeight: 34)
            .background(RoundedRectangle(cornerRadius: Brand.radius).fill(Brand.gold))
            .opacity(enabled ? (configuration.isPressed ? 0.8 : 1) : 0.4)
            .scaleEffect(configuration.isPressed ? 0.98 : 1)
    }
}

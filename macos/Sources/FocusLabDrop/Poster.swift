import AppKit
import CoreImage.CIFilterBuiltins
import PDFKit
import SwiftUI

/// Black on white QR with whole-pixel modules so it stays sharp when printed.
func qrImage(for text: String) -> NSImage? {
    let filter = CIFilter.qrCodeGenerator()
    filter.message = Data(text.utf8)
    filter.correctionLevel = "M"
    guard let out = filter.outputImage else { return nil }
    let scale = (1000 / out.extent.width).rounded(.up)
    let big = out.samplingNearest().transformed(by: CGAffineTransform(scaleX: scale, y: scale))
    guard let cg = CIContext().createCGImage(big, from: big.extent) else { return nil }
    return NSImage(cgImage: cg, size: NSSize(width: big.extent.width, height: big.extent.height))
}

/// A4 portrait (595 x 842 pt). Only the upload link; the Wi-Fi QR is managed by hand.
struct PosterView: View {
    let url: String
    private let ink = Color(red: 13 / 255, green: 27 / 255, blue: 42 / 255)
    private let slate = Color(red: 71 / 255, green: 85 / 255, blue: 105 / 255)

    var body: some View {
        VStack(spacing: 0) {
            if let logo = Brand.logo {
                Image(nsImage: logo).resizable().frame(width: 80, height: 80)
            }
            Text("Focus Digital Color Lab")
                .font(.system(size: 20, weight: .semibold)).foregroundStyle(ink).padding(.top, 10)

            Text("Send your photos\nfor printing")
                .font(.system(size: 40, weight: .bold)).multilineTextAlignment(.center)
                .fixedSize(horizontal: false, vertical: true)
                .foregroundStyle(ink).padding(.top, 34)
            Text("Point your phone camera at the code. Your photos go straight to our print desk at full quality.")
                .font(.system(size: 16)).multilineTextAlignment(.center)
                .fixedSize(horizontal: false, vertical: true)
                .foregroundStyle(slate).frame(width: 400).padding(.top, 12)

            Group {
                if let qr = qrImage(for: url) {
                    Image(nsImage: qr).resizable().interpolation(.none).frame(width: 270, height: 270)
                }
            }
            .padding(18)
            .background(RoundedRectangle(cornerRadius: 16).fill(Color.white))
            .overlay(RoundedRectangle(cornerRadius: 16).stroke(Brand.gold, lineWidth: 5))
            .padding(.top, 30)

            Text("Not connected yet? Join our Wi-Fi first, then scan again.")
                .font(.system(size: 15, weight: .medium)).foregroundStyle(ink).padding(.top, 26)
            Text(url)
                .font(.system(size: 12, design: .monospaced)).foregroundStyle(slate).padding(.top, 8)

            Spacer(minLength: 14)
            Text("Photos stay on our counter computer. Nothing is sent over the internet.")
                .font(.system(size: 12)).foregroundStyle(slate)
        }
        .padding(.top, 40).padding(.bottom, 36)
        .frame(width: 595, height: 842)
        .background(Color.white)
    }
}

@MainActor
func posterPDF(url: String) -> Data? {
    let renderer = ImageRenderer(content: PosterView(url: url))
    let data = NSMutableData()
    renderer.render { size, draw in
        var box = CGRect(origin: .zero, size: size)
        guard let consumer = CGDataConsumer(data: data as CFMutableData),
              let ctx = CGContext(consumer: consumer, mediaBox: &box, nil) else { return }
        ctx.beginPDFPage(nil)
        draw(ctx)
        ctx.endPDFPage()
        ctx.closePDF()
    }
    return data.length > 0 ? data as Data : nil
}

struct PosterPreview: View {
    let url: String

    var body: some View {
        VStack(spacing: 16) {
            PosterView(url: url)
                .scaleEffect(0.62, anchor: .center)
                .frame(width: 595 * 0.62, height: 842 * 0.62)
                .shadow(color: .black.opacity(0.2), radius: 8, y: 2)
            HStack {
                Button("Save as PDF...") { savePDF() }
                Spacer()
                Button("Print...") { printPoster() }.buttonStyle(GoldButtonStyle())
            }
        }
        .padding(20)
        .frame(width: 595 * 0.62 + 40)
    }

    private func savePDF() {
        guard let data = posterPDF(url: url) else { return AppModel.shared.alert("Could not create the PDF.") }
        let panel = NSSavePanel()
        panel.nameFieldStringValue = "FocusLab upload QR.pdf"
        panel.allowedContentTypes = [.pdf]
        guard panel.runModal() == .OK, let target = panel.url else { return }
        do { try data.write(to: target) } catch { AppModel.shared.alert("Could not save the PDF: \(error.localizedDescription)") }
    }

    private func printPoster() {
        guard let data = posterPDF(url: url), let doc = PDFDocument(data: data) else {
            return AppModel.shared.alert("Could not prepare the poster for printing.")
        }
        doc.printOperation(for: .shared, scalingMode: .pageScaleToFit, autoRotate: true)?.run()
    }
}

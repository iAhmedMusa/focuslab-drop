import Darwin
import Foundation

public struct LANAddress: Equatable, Hashable, Identifiable {
    public let interface: String
    public let ip: String
    public var id: String { ip }
}

public func isPrivateIPv4(_ ip: String) -> Bool {
    let p = ip.split(separator: ".", omittingEmptySubsequences: false).compactMap { Int($0) }
    guard p.count == 4, p.allSatisfy({ (0...255).contains($0) }) else { return false }
    switch (p[0], p[1]) {
    case (10, _), (192, 168), (172, 16...31): return true
    default: return false
    }
}

/// Wi-Fi and Ethernet addresses (en*) that a phone on the shop network could reach.
public func lanAddresses() -> [LANAddress] {
    var head: UnsafeMutablePointer<ifaddrs>?
    guard getifaddrs(&head) == 0, let first = head else { return [] }
    defer { freeifaddrs(head) }
    var found: [LANAddress] = []
    for ptr in sequence(first: first, next: { $0.pointee.ifa_next }) {
        let ifa = ptr.pointee
        guard let sa = ifa.ifa_addr, sa.pointee.sa_family == UInt8(AF_INET) else { continue }
        let flags = Int32(ifa.ifa_flags)
        guard flags & IFF_UP != 0, flags & IFF_LOOPBACK == 0 else { continue }
        var host = [CChar](repeating: 0, count: Int(NI_MAXHOST))
        guard getnameinfo(sa, socklen_t(sa.pointee.sa_len), &host, socklen_t(host.count), nil, 0, NI_NUMERICHOST) == 0 else { continue }
        let ip = String(cString: host)
        let name = String(cString: ifa.ifa_name)
        if name.hasPrefix("en"), isPrivateIPv4(ip) { found.append(LANAddress(interface: name, ip: ip)) }
    }
    return found.sorted { $0.interface < $1.interface }
}

public func customerURL(ip: String, port: Int) -> String { "http://\(ip):\(port)/drop" }
public func counterURL(port: Int) -> String { "http://localhost:\(port)/lab-counter" }
public func isValidPort(_ port: Int) -> Bool { (1024...65535).contains(port) }

/// `-exit-on-stdin-close` makes the server quit if the app dies and its stdin pipe closes.
public func serverArguments(dir: String, port: Int) -> [String] {
    ["-dir", dir, "-port", String(port), "-exit-on-stdin-close"]
}

/// True if something already accepts connections on 127.0.0.1:port.
public func isPortInUse(_ port: Int) -> Bool {
    let fd = socket(AF_INET, SOCK_STREAM, 0)
    guard fd >= 0 else { return false }
    defer { close(fd) }
    var addr = sockaddr_in()
    addr.sin_len = UInt8(MemoryLayout<sockaddr_in>.size)
    addr.sin_family = sa_family_t(AF_INET)
    addr.sin_port = in_port_t(UInt16(truncatingIfNeeded: port)).bigEndian
    addr.sin_addr.s_addr = inet_addr("127.0.0.1")
    return withUnsafePointer(to: &addr) {
        $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
            connect(fd, $0, socklen_t(MemoryLayout<sockaddr_in>.size)) == 0
        }
    }
}

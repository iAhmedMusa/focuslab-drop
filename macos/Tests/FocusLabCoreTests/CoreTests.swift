import Darwin
import XCTest
@testable import FocusLabCore

final class CoreTests: XCTestCase {
    func testPrivateAddresses() {
        for ip in ["192.168.0.101", "10.0.0.5", "172.16.0.1", "172.31.255.254"] { XCTAssertTrue(isPrivateIPv4(ip), ip) }
        for ip in ["8.8.8.8", "172.32.0.1", "172.15.0.1", "127.0.0.1", "169.254.1.1", "192.168.0", "192.168.0.256", "", "abc"] {
            XCTAssertFalse(isPrivateIPv4(ip), ip)
        }
    }

    func testURLsAndPorts() {
        XCTAssertEqual(customerURL(ip: "192.168.1.50", port: 6789), "http://192.168.1.50:6789/drop")
        XCTAssertEqual(counterURL(port: 7000), "http://localhost:7000/lab-counter")
        XCTAssertTrue(isValidPort(6789))
        XCTAssertFalse(isValidPort(80))
        XCTAssertFalse(isValidPort(70000))
    }

    func testServerArgumentsAlwaysCarryTheExitFlag() {
        XCTAssertEqual(serverArguments(dir: "/Users/shop/My Prints", port: 6789),
                       ["-dir", "/Users/shop/My Prints", "-port", "6789", "-exit-on-stdin-close"])
    }

    func testLanAddressesArePrivateAndNotLoopback() {
        for a in lanAddresses() {
            XCTAssertTrue(isPrivateIPv4(a.ip))
            XCTAssertTrue(a.interface.hasPrefix("en"))
        }
    }

    func testPortInUseDetection() throws {
        let fd = socket(AF_INET, SOCK_STREAM, 0)
        XCTAssertGreaterThanOrEqual(fd, 0)
        defer { close(fd) }
        var addr = sockaddr_in()
        addr.sin_len = UInt8(MemoryLayout<sockaddr_in>.size)
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_addr.s_addr = inet_addr("127.0.0.1") // port 0: the system picks a free one
        let size = socklen_t(MemoryLayout<sockaddr_in>.size)
        XCTAssertEqual(withUnsafePointer(to: &addr) { $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { Darwin.bind(fd, $0, size) } }, 0)
        XCTAssertEqual(Darwin.listen(fd, 1), 0)
        var bound = sockaddr_in()
        var len = size
        XCTAssertEqual(withUnsafeMutablePointer(to: &bound) { $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { getsockname(fd, $0, &len) } }, 0)
        let port = Int(UInt16(bigEndian: bound.sin_port))

        XCTAssertTrue(isPortInUse(port))
        close(fd)
        XCTAssertFalse(isPortInUse(port))
    }
}

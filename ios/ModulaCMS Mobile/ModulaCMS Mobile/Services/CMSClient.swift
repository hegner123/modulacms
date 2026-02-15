import Foundation
import ModulaCMS

enum CMSClient {
    static var shared: ModulaCMSClient = makeClient()

    static func reconfigure() {
        shared = makeClient()
    }

    private static func makeClient() -> ModulaCMSClient {
        let config = ClientConfig(
            baseURL: AppConfig.baseURL,
            apiKey: AppConfig.apiKey
        )
        do {
            return try ModulaCMSClient(config: config)
        } catch {
            fatalError("Failed to create ModulaCMS client: \(error)")
        }
    }
}

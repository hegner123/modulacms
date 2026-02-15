import Foundation
import ModulaCMS

@Observable
final class SettingsViewModel {
    var baseURL: String = AppConfig.baseURL
    var apiKey: String = AppConfig.apiKey
    var authorID: String = AppConfig.authorID
    var testState: LoadState<String> = .idle
    var showSaveConfirmation = false

    var hasChanges: Bool {
        baseURL != AppConfig.baseURL || apiKey != AppConfig.apiKey || authorID != AppConfig.authorID
    }

    func save() {
        AppConfig.baseURL = baseURL
        AppConfig.apiKey = apiKey
        AppConfig.authorID = authorID
        CMSClient.reconfigure()
        showSaveConfirmation = true
    }

    func testConnection() async {
        testState = .loading
        do {
            let config = ClientConfig(baseURL: baseURL, apiKey: apiKey)
            let client = try ModulaCMSClient(config: config)
            let routes = try await client.routes.list()
            testState = .loaded("Connected — \(routes.count) route(s) found")
        } catch {
            testState = .error(error.localizedDescription)
        }
    }
}

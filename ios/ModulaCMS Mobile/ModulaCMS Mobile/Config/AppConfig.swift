import Foundation

enum AppConfig {
    private static let defaults = UserDefaults.standard

    private enum Keys {
        static let baseURL = "cms_base_url"
        static let apiKey = "cms_api_key"
        static let authorID = "cms_author_id"
    }

    static var baseURL: String {
        get { defaults.string(forKey: Keys.baseURL) ?? "http://localhost:8080" }
        set { defaults.set(newValue, forKey: Keys.baseURL) }
    }

    static var apiKey: String {
        get { defaults.string(forKey: Keys.apiKey) ?? "dev-api-key" }
        set { defaults.set(newValue, forKey: Keys.apiKey) }
    }

    static var authorID: String {
        get { defaults.string(forKey: Keys.authorID) ?? "" }
        set { defaults.set(newValue, forKey: Keys.authorID) }
    }

    static var isConfigured: Bool {
        !baseURL.isEmpty
    }

    static var needsSetup: Bool {
        defaults.string(forKey: Keys.apiKey) == nil
    }
}

import Foundation

struct SelectOptions: Codable {
    let options: [String]
}

enum FieldConfig {
    static func parseSelectOptions(from data: String) -> [String] {
        guard let jsonData = data.data(using: .utf8) else { return [] }
        let decoded = try? JSONDecoder().decode(SelectOptions.self, from: jsonData)
        return decoded?.options ?? []
    }
}

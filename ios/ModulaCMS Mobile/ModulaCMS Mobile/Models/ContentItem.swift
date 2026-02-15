import Foundation
import ModulaCMS

struct ContentItem: Identifiable {
    let contentData: ContentData
    let route: Route?
    let datatypeLabel: String

    var id: String { contentData.contentDataID.rawValue }
    var title: String { route?.title ?? "Untitled" }
    var slug: String { route?.slug.rawValue ?? "" }
    var status: ContentStatus { contentData.status }
    var dateModified: Timestamp { contentData.dateModified }
}

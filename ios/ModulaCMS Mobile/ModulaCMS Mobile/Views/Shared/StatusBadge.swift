import SwiftUI
import ModulaCMS

struct StatusBadge: View {
    let status: ContentStatus

    var body: some View {
        Text(status.rawValue.capitalized)
            .font(.caption2)
            .fontWeight(.medium)
            .padding(.horizontal, 8)
            .padding(.vertical, 4)
            .background(color.opacity(0.15))
            .foregroundStyle(color)
            .clipShape(Capsule())
    }

    private var color: Color {
        switch status {
        case .published: .green
        case .draft: .orange
        case .archived: .gray
        case .pending: .blue
        }
    }
}

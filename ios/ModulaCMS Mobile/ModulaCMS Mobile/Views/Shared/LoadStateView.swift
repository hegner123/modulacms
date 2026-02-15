import SwiftUI

struct LoadStateView<T, Content: View>: View {
    let state: LoadState<T>
    let emptyMessage: String
    @ViewBuilder let content: (T) -> Content

    var body: some View {
        switch state {
        case .idle, .loading:
            ProgressView()
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        case .loaded(let data):
            if isEmpty(data) {
                ContentUnavailableView(emptyMessage, systemImage: "tray")
            } else {
                content(data)
            }
        case .error(let message):
            ContentUnavailableView {
                Label("Error", systemImage: "exclamationmark.triangle")
            } description: {
                Text(message)
            }
        }
    }

    private func isEmpty(_ data: T) -> Bool {
        if let array = data as? (any Collection) {
            return array.isEmpty
        }
        return false
    }
}

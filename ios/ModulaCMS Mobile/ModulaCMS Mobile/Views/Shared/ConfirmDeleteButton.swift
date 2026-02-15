import SwiftUI

struct ConfirmDeleteButton: View {
    let title: String
    let message: String
    let action: () -> Void

    @State private var showConfirmation = false

    var body: some View {
        Button(role: .destructive) {
            showConfirmation = true
        } label: {
            Label("Delete", systemImage: "trash")
        }
        .confirmationDialog(title, isPresented: $showConfirmation, titleVisibility: .visible) {
            Button("Delete", role: .destructive, action: action)
            Button("Cancel", role: .cancel) {}
        } message: {
            Text(message)
        }
    }
}

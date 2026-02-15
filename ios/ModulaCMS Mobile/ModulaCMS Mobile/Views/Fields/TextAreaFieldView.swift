import SwiftUI

struct TextAreaFieldView: View {
    let label: String
    @Binding var value: String

    var body: some View {
        TextField(label, text: $value, axis: .vertical)
            .lineLimit(3...8)
    }
}

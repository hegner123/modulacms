import SwiftUI

struct TextFieldView: View {
    let label: String
    @Binding var value: String

    var body: some View {
        TextField(label, text: $value)
    }
}

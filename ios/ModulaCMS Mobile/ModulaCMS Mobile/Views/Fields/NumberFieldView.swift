import SwiftUI

struct NumberFieldView: View {
    let label: String
    @Binding var value: String

    var body: some View {
        TextField(label, text: $value)
            .keyboardType(.decimalPad)
    }
}

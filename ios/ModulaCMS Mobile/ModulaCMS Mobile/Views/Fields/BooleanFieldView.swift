import SwiftUI

struct BooleanFieldView: View {
    let label: String
    @Binding var value: String

    private var isOn: Binding<Bool> {
        Binding(
            get: { value.lowercased() == "true" },
            set: { value = $0 ? "true" : "false" }
        )
    }

    var body: some View {
        Toggle(label, isOn: isOn)
    }
}

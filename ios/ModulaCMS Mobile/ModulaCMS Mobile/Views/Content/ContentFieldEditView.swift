import SwiftUI
import ModulaCMS

struct ContentFieldEditView: View {
    let fieldPairs: [FieldPair]
    @Binding var values: [String: String]

    var body: some View {
        ForEach(fieldPairs) { pair in
            FieldRenderer(
                field: pair.field,
                value: binding(for: pair.field.fieldID.rawValue),
                readOnly: false
            )
        }
    }

    private func binding(for key: String) -> Binding<String> {
        Binding(
            get: { values[key] ?? "" },
            set: { values[key] = $0 }
        )
    }
}

import SwiftUI
import ModulaCMS

struct ContentFieldsSection: View {
    let fieldPairs: [FieldPair]

    var body: some View {
        ForEach(fieldPairs) { pair in
            FieldRenderer(
                field: pair.field,
                value: .constant(pair.value),
                readOnly: true
            )
        }
    }
}

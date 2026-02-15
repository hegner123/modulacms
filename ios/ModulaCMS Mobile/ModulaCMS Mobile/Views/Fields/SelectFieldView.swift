import SwiftUI

struct SelectFieldView: View {
    let label: String
    @Binding var value: String
    let options: [String]

    var body: some View {
        Picker(label, selection: $value) {
            Text("None").tag("")
            ForEach(options, id: \.self) { option in
                Text(option).tag(option)
            }
        }
    }
}

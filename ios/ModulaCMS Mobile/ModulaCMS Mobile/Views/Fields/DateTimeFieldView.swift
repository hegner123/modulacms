import SwiftUI

struct DateTimeFieldView: View {
    let label: String
    @Binding var value: String

    @State private var date = Date()
    @State private var initialized = false

    var body: some View {
        DatePicker(label, selection: $date, displayedComponents: [.date, .hourAndMinute])
            .onAppear {
                if !initialized {
                    if let parsed = ISO8601DateFormatter().date(from: value) {
                        date = parsed
                    }
                    initialized = true
                }
            }
            .onChange(of: date) { _, newDate in
                value = ISO8601DateFormatter().string(from: newDate)
            }
    }
}

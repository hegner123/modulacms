import SwiftUI
import ModulaCMS

struct FieldRenderer: View {
    let field: Field
    @Binding var value: String
    let readOnly: Bool

    var body: some View {
        if readOnly {
            readOnlyView
        } else {
            editableView
        }
    }

    @ViewBuilder
    private var readOnlyView: some View {
        switch field.type {
        case .boolean:
            LabeledContent(field.label) {
                Image(systemName: value.lowercased() == "true" ? "checkmark.circle.fill" : "xmark.circle")
                    .foregroundStyle(value.lowercased() == "true" ? .green : .secondary)
            }
        default:
            LabeledContent(field.label, value: value.isEmpty ? "--" : value)
        }
    }

    @ViewBuilder
    private var editableView: some View {
        switch field.type {
        case .text:
            TextFieldView(label: field.label, value: $value)
        case .textarea, .richtext:
            TextAreaFieldView(label: field.label, value: $value)
        case .number:
            NumberFieldView(label: field.label, value: $value)
        case .boolean:
            BooleanFieldView(label: field.label, value: $value)
        case .select:
            SelectFieldView(
                label: field.label,
                value: $value,
                options: FieldConfig.parseSelectOptions(from: field.data)
            )
        case .date:
            DateFieldView(label: field.label, value: $value)
        case .datetime:
            DateTimeFieldView(label: field.label, value: $value)
        default:
            TextFieldView(label: field.label, value: $value)
        }
    }
}

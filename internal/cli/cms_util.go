package cli

import (
	"github.com/hegner123/modulacms/internal/db/types"
)

// CollectFieldValuesFromForm extracts field values from form state
// Returns map[field_id]field_value
func (m Model) CollectFieldValuesFromForm() map[types.FieldID]string {
	fieldValues := make(map[types.FieldID]string)

	// FormState.FormValues is []*string and FormState.FormMap is []string
	// where FormMap contains field_id as string
	for i, value := range m.FormState.FormValues {
		if value == nil || *value == "" {
			continue
		}

		// Parse field ID from FormMap
		if i < len(m.FormState.FormMap) {
			fieldIDStr := m.FormState.FormMap[i]
			fieldID := types.FieldID(fieldIDStr)
			if err := fieldID.Validate(); err != nil {
				m.Logger.Ferror("Invalid field ID", err)
				continue
			}

			fieldValues[fieldID] = *value
		}
	}

	return fieldValues
}

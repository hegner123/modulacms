package cli

import (
	"strconv"

	"github.com/hegner123/modulacms/internal/utility"
)

// CollectFieldValuesFromForm extracts field values from form state
// Returns map[field_id]field_value
func (m Model) CollectFieldValuesFromForm() map[int64]string {
	fieldValues := make(map[int64]string)

	// FormState.FormValues is []*string and FormState.FormMap is []string
	// where FormMap contains field_id as string
	for i, value := range m.FormState.FormValues {
		if value == nil || *value == "" {
			continue
		}

		// Parse field ID from FormMap
		if i < len(m.FormState.FormMap) {
			fieldIDStr := m.FormState.FormMap[i]
			fieldID, err := strconv.ParseInt(fieldIDStr, 10, 64)
			if err != nil {
				utility.DefaultLogger.Ferror("Failed to parse field ID", err)
				continue
			}

			fieldValues[fieldID] = *value
		}
	}

	return fieldValues
}

package definitions

import "fmt"

// Validate checks a SchemaDefinition for internal consistency.
func Validate(def SchemaDefinition) error {
	if def.Name == "" {
		return fmt.Errorf("definitions: name cannot be empty")
	}

	if len(def.Datatypes) == 0 {
		return fmt.Errorf("definitions: %q must have at least one datatype", def.Name)
	}

	// Validate each datatype
	for key, dt := range def.Datatypes {
		if dt.Label == "" {
			return fmt.Errorf("definitions: %q datatype %q has empty label", def.Name, key)
		}
		if !dt.Type.Valid || dt.Type.String == "" {
			return fmt.Errorf("definitions: %q datatype %q has empty type", def.Name, key)
		}

		// Validate ParentRef references an existing datatype (not self)
		if dt.ParentRef != "" {
			if dt.ParentRef == key {
				return fmt.Errorf("definitions: %q datatype %q has self-reference in ParentRef", def.Name, key)
			}
			if _, ok := def.Datatypes[dt.ParentRef]; !ok {
				return fmt.Errorf("definitions: %q datatype %q references unknown parent %q", def.Name, key, dt.ParentRef)
			}
		}

		// Validate inline fields
		for i, field := range dt.FieldRefs {
			if field.Label == "" {
				return fmt.Errorf("definitions: %q datatype %q field[%d] has empty label", def.Name, key, i)
			}
			if err := field.Type.Validate(); err != nil {
				return fmt.Errorf("definitions: %q datatype %q field[%d] %q has invalid type: %w", def.Name, key, i, field.Label, err)
			}
		}
	}

	return nil
}

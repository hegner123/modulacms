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

	if len(def.RootKeys) == 0 {
		return fmt.Errorf("definitions: %q must have at least one root key", def.Name)
	}

	// Validate RootKeys reference existing datatypes
	for _, key := range def.RootKeys {
		if _, ok := def.Datatypes[key]; !ok {
			return fmt.Errorf("definitions: %q root key %q not found in datatypes", def.Name, key)
		}
	}

	// Validate each datatype
	for key, dt := range def.Datatypes {
		if dt.Label == "" {
			return fmt.Errorf("definitions: %q datatype %q has empty label", def.Name, key)
		}
		if dt.Type == "" {
			return fmt.Errorf("definitions: %q datatype %q has empty type", def.Name, key)
		}

		// Validate FieldRefs reference existing fields
		for _, fieldRef := range dt.FieldRefs {
			if _, ok := def.Fields[fieldRef]; !ok {
				return fmt.Errorf("definitions: %q datatype %q references unknown field %q", def.Name, key, fieldRef)
			}
		}

		// Validate ChildRefs reference existing datatypes (no self-references)
		for _, childRef := range dt.ChildRefs {
			if childRef == key {
				return fmt.Errorf("definitions: %q datatype %q has self-reference in ChildRefs", def.Name, key)
			}
			if _, ok := def.Datatypes[childRef]; !ok {
				return fmt.Errorf("definitions: %q datatype %q references unknown child datatype %q", def.Name, key, childRef)
			}
		}
	}

	// Validate each field
	for key, field := range def.Fields {
		if field.Label == "" {
			return fmt.Errorf("definitions: %q field %q has empty label", def.Name, key)
		}
		if err := field.Type.Validate(); err != nil {
			return fmt.Errorf("definitions: %q field %q has invalid type: %w", def.Name, key, err)
		}
	}

	return nil
}

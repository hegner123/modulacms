package config

// MergeMaps performs a shallow merge of overlay onto base. Keys present in
// overlay overwrite the corresponding keys in base. Keys in base that are
// absent from overlay are preserved. Maps and slices are replaced entirely
// (not recursively merged).
//
// Neither input map is modified; the result is a new map.
func MergeMaps(base, overlay map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(overlay))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range overlay {
		merged[k] = v
	}
	return merged
}

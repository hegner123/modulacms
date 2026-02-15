package cli

import "strings"

// MakeFilter returns the provided arguments as a slice.
func MakeFilter(args ...string)[]string{
        return args
}

// GetLine returns the first string in the slice that contains the match string.
func GetLine(s []string, match string) string {
	for _, v := range s {
		if strings.Contains(v, match) {
			return v
		}
	}
	return ""
}

// FilterLines removes strings from the slice that contain any of the remove strings.
func FilterLines(s *[]string, remove []string) {
	result := make([]string, 0)
	for _, v := range *s {
		shouldRemove := false
		for _, rm := range remove {
			if strings.Contains(v, rm) {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			result = append(result, v)
		}
	}
	*s = result
}

package cli

import "strings"

func MakeFilter(args ...string)[]string{
        return args
}

func GetLine(s []string, match string) string {
	for _, v := range s {
		if strings.Contains(v, match) {
			return v
		}
	}
	return ""
}

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

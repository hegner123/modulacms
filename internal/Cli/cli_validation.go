package cli

import "fmt"

func Required(s string) error {
	if len(s) < 1 {
		return fmt.Errorf("\nInput Cannot Be Null")
	} else {
		return nil
	}

}

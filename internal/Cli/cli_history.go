package cli

import (
	"fmt"
	"io"
	"os"

	utility "github.com/hegner123/modulacms/internal/Utility"
)

func AddToHistory(p CliPage) error {
	file, err := os.OpenFile("history.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening or creating file:", err)
		return err
	}
	defer file.Close()

	h, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	history := string(h)
	entry := fmt.Sprintf("%s %d %s\n", utility.TimestampS(), p.Index, p.Label)
	entry += history
	_, err = file.WriteString(entry)
	if err != nil {
		return err
	}
	return nil
}

func GetPrevHistory(index int) (string, int) {
	file, err := os.OpenFile("history.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening or creating file:", err)
		return "err", 0
	}
	defer file.Close()
	return "", 0
}

func GetNextHistory(index int) (string, int) {
	file, err := os.OpenFile("history.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening or creating file:", err)
		return "err", 0
	}
	defer file.Close()
	return "", 0
}

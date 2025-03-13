package db

import (
	"encoding/json"
	"fmt"
)

func MapHistory[T Historied](entry T) (string, error) {
	data := []byte(entry.GetHistory())
	e := []byte(entry.MapHistoryEntry())

	var entries []map[string]any
	if err := json.Unmarshal(data, &entries); err != nil {
		return "", err
	}

	var newEntry map[string]any
	if err := json.Unmarshal(e, &newEntry); err != nil {
		return "", err
	}

	entries = append(entries, newEntry)

	result, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func PopHistory[T Historied](entry T) (*Historied, error) {
	data := []byte(entry.GetHistory())

	var entries []map[string]any
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	p := entries[len(entries)-1]
	h := entries[:len(entries)-1]
    


    fmt.Printf("last entry: %v\n", p)
    fmt.Printf("remaining entries: %v\n", h)
	return nil, nil

}

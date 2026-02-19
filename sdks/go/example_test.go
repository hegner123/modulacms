package modula

import (
	"fmt"
	"time"
)

func ExampleNewClient() {
	client, err := NewClient(ClientConfig{
		BaseURL: "https://cms.example.com",
		APIKey:  "test-key",
	})
	if err != nil {
		panic(err)
	}
	_ = client
	// Output:
}

func ExampleContentID() {
	id := ContentID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	fmt.Println(id.String())
	fmt.Println(id.IsZero())
	// Output:
	// 01ARZ3NDEKTSV4RRFFQ69G5FAV
	// false
}

func ExampleTimestamp() {
	ts := NewTimestamp(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	fmt.Println(ts.String())
	// Output:
	// 2024-01-15T10:30:00Z
}

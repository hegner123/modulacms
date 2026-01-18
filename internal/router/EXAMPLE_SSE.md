```go
func sseHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers required for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Ensure the connection supports streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Send an initial message: Process launched.
	fmt.Fprintf(w, "data: Process launched\n\n")
	flusher.Flush()

	// Launch the long-running process in a background goroutine.
	done := make(chan string)
	go func() {
		// Simulate a long process (e.g., processing a file, making external API calls, etc.)
		time.Sleep(5 * time.Second) // Simulated delay
		// Signal that the process has completed.
		done <- "Process completed successfully!"
	}()

	// Wait for the background process to complete and then send a final update.
	result := <-done
	fmt.Fprintf(w, "data: %s\n\n", result)
	flusher.Flush()

	// Optionally, you can keep the connection open to send more updates or close it.
}

package certificates



import (
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func SSLCertManager() {
	// Create an autocert.Manager instance
	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"), // Folder to store certificates
		HostPolicy: autocert.HostWhitelist("example.com"), // Your domain(s)
	}

	// Configure an HTTPS server that uses the TLS configuration provided by autocert
	server := &http.Server{
		Addr:      ":https",
		TLSConfig: manager.TLSConfig(),
		Handler:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, HTTPS world!"))
		}),
	}

	// Start a HTTP server that will redirect HTTP to HTTPS and allow Let's Encrypt's ACME challenge responses.
	go func() {
		httpServer := &http.Server{
			Addr:    ":http",
			Handler: manager.HTTPHandler(nil),
		}
		log.Fatal(httpServer.ListenAndServe())
	}()

	// Start the HTTPS server; ListenAndServeTLS uses empty strings because TLSConfig is provided by autocert.
	log.Fatal(server.ListenAndServeTLS("", ""))
}


package server

// DefaultPort is the HTTP port used when no override is provided in configuration.
const DefaultPort = "8080"

// ReadTimeout is the maximum duration for reading a complete request, including body.
const ReadTimeout = "15s"

// WriteTimeout is the maximum duration before timing out writes of a response.
const WriteTimeout = "15s"

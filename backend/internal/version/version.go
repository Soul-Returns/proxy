package version

// Current is the current version of the backend (set via ldflags during build)
var Current = "1.0.0"

// GetVersion returns the current backend version
func GetVersion() string {
	return Current
}

package version

// Values for these are injected by the build
var (
	version = "devel"
	commit  string
)

// Version returns the brignext version. This is typically a semantic version,
// but in the case of unreleased code, could be another descriptor such as
// "edge".
func Version() string {
	return version
}

// Commit returns the git commit SHA for the code that brignext was built from.
func Commit() string {
	return commit
}

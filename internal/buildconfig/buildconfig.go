package buildconfig

const (
	snapshot = "snapshot"
)

var (
	version = snapshot
	commit  string
)

func Version() string {
	return version
}

func Commit() string {
	return commit
}

func IsRelease() bool {
	return !IsDevelopment()
}

func IsDevelopment() bool {
	return version == snapshot
}

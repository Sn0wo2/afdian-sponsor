package version

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func GetVersion() string {
	return version
}

func GetCommit() string {
	return commit
}

func GetDate() string {
	return date
}

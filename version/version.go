package version

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func String() string {
	return version + "-" + commit + "(" + date + ")"
}

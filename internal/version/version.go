package version

import "fmt"

var (
	Version   = "dev"
	BuildDate = "unknown"

	BuildHash = "unknown"
)

func String() string {
	return fmt.Sprintf("build %s (%s) · %s", Version, BuildHash, BuildDate)
}

package version

import "fmt"

var (
	Version   = "dev"
	BuildDate = "unknown"

	BuildHash = "unknown"
)

func String() string {
	return fmt.Sprintf("build %s (%s) · %s", Short(), BuildHash, BuildDate)
}

func Short() string {
	if len(Version) > 7 {
		return Version[:7]
	}
	return Version
}

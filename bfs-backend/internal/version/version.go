package version

import "os"

var Version = "dev"

func init() {
	if v := os.Getenv("APP_VERSION"); v != "" {
		Version = v
	}
}

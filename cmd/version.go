package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "latest"
	Revision  = "HEAD"
	Branch    = "main"
	BuildUser = "(nobody)"
	BuildDate = "(unknown)"
	GoVersion = runtime.Version()
	Platform  = runtime.GOOS + "/" + runtime.GOARCH
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long:  `Print version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(VersionInfo())
	},
}

func VersionInfo() (versionInfo string) {
	versionInfo = fmt.Sprintf(
		"kube-sealer %s (revision=%s, branch=%s, buildUser=%s, buildDate=%s, go=%s, platform=%s)",
		Version, Revision, Branch, BuildUser, BuildDate, GoVersion, Platform,
	)
	return versionInfo
}

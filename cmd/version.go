// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// buildVersion function (no changes)
func buildVersion(Version, Commit, Date, BuiltBy string) string {
	result := Version
	if Commit != "" {
		result = fmt.Sprintf("%s\nCommit: %s\n", result, Commit)
	}
	if Date != "" {
		result = fmt.Sprintf("%sBuilt at: %s\n", result, Date)
	}
	if BuiltBy != "" {
		result = fmt.Sprintf("%sBuilt by: %s\n", result, BuiltBy)
	}
	result = fmt.Sprintf(
		"%sGOOS: %s\nGOARCH: %s\n", result, runtime.GOOS, runtime.GOARCH,
	)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf(
			"%smodule Version: %s, checksum: %s",
			result,
			info.Main.Version,
			info.Main.Sum,
		)
	}
	return result
}

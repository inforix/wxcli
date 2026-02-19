package cmd

import "runtime/debug"

func VersionString() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if info.Main.Version == "" {
		return "dev"
	}
	return info.Main.Version
}

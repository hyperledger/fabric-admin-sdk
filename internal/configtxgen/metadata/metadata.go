/*
Copyright 2017 Hitachi America

SPDX-License-Identifier: Apache-2.0
*/

package metadata

import (
	"fmt"
	"runtime"
)

const ProgramName = "configtxgen"

var (
	CommitSHA = "development build"
	Version   = "latest"
)

func GetVersionInfo() string {
	return fmt.Sprintf("%s:\n Version: %s\n Commit SHA: %s\n Go version: %s\n OS/Arch: %s",
		ProgramName, Version, CommitSHA, runtime.Version(),
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
}

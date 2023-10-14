package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const ProgramName = "peer"

var (
	Version   string = "latest"
	CommitSHA string = "development build"
)

// Cmd returns the Cobra Command for Version
func Cmd() *cobra.Command {
	return cobraCommand
}

var cobraCommand = &cobra.Command{
	Use:   "version",
	Short: "Print fabric peer version.",
	Long:  `Print current version of the fabric peer server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("trailing args detected")
		}
		// Parsing of the command line is done so silence cmd usage
		cmd.SilenceUsage = true
		fmt.Print(GetInfo())
		return nil
	},
}

// GetInfo returns version information for the peer
func GetInfo() string {
	return fmt.Sprintf("%s:\n Version: %s\n Commit SHA: %s\n Go version: %s\n"+
		" OS/Arch: %s\n",
		ProgramName, Version, CommitSHA, runtime.Version(),
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
}

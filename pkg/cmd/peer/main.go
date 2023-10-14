package peer

import (
	"os"

	"github.com/hyperledger/fabric-admin-sdk/pkg/cmd/peer/lifecycle"
	"github.com/hyperledger/fabric-admin-sdk/pkg/cmd/peer/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mainCmd = &cobra.Command{Use: "peer"}

func main() {
	// Define command-line flags that are valid for all peer commands and
	// subcommands.
	mainFlags := mainCmd.PersistentFlags()

	mainFlags.String("logging-level", "", "Legacy logging level flag")
	viper.BindPFlag("logging_level", mainFlags.Lookup("logging-level"))
	mainFlags.MarkHidden("logging-level")

	mainCmd.AddCommand(version.Cmd())
	mainCmd.AddCommand(lifecycle.Cmd())

	// On failure Cobra prints the usage message and error string, so we only
	// need to exit with a non-0 status
	if mainCmd.Execute() != nil {
		os.Exit(1)
	}
}

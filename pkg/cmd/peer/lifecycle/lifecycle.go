package lifecycle

import (
	"github.com/spf13/cobra"
)

var (
	mspID          string
	SignCert       string
	PrivKeyPath    string
	orgPeerAddress string
	TLSCACert      string
)

// Cmd returns the cobra command for lifecycle
func Cmd() *cobra.Command {
	lifecycleCmd := &cobra.Command{
		Use:   "lifecycle",
		Short: "Perform _lifecycle operations",
		Long:  "Perform _lifecycle operations",
	}
	addPeerCommonFlags(lifecycleCmd)
	lifecycleCmd.AddCommand(InstallCmd())
	return lifecycleCmd
}

func addPeerCommonFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.StringVarP(&mspID, "msp", "", "",
		"msp id for your msp")
	flags.StringVarP(&SignCert, "SignCert", "", "",
		"Sign Cert for your msp")
	flags.StringVarP(&PrivKeyPath, "PrivKeyPath", "", "",
		"Priv Key Path for your msp")
	flags.StringVarP(&orgPeerAddress, "PeerAddress", "", "",
		"Target Peer address")
	flags.StringVarP(&TLSCACert, "TLSCACert", "", "",
		"TLS CA Cert to connect to peer")
}

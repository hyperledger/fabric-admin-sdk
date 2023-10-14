package lifecycle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	"github.com/hyperledger/fabric-admin-sdk/pkg/cmd/peer/common"
	"github.com/spf13/cobra"
)

var packageFilePath string

func InstallCmd() *cobra.Command {
	chaincodeInstallCmd := &cobra.Command{
		Use:       "install",
		Short:     "Install a chaincode.",
		Long:      "Install a chaincode on a peer.",
		ValidArgs: []string{"1"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			ConnectionDetail, err := common.ConstructConnectionDetail(mspID, SignCert, PrivKeyPath, orgPeerAddress, TLSCACert)
			if err != nil {
				return err
			}
			packageReader, err := os.Open(packageFilePath)
			if err != nil {
				return err
			}
			chaincodePackage, err := io.ReadAll(packageReader)
			if err != nil {
				return err
			}
			result, err := chaincode.Install(ctx, ConnectionDetail.Connection, ConnectionDetail.ID, bytes.NewReader(chaincodePackage))
			if err != nil {
				return err
			}
			fmt.Printf("Install success, package id [%s], package label [%s] \n", result.GetPackageId(), result.GetLabel())
			return nil
		},
	}
	addInstallFlags(chaincodeInstallCmd)
	return chaincodeInstallCmd
}

func addInstallFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.StringVarP(&packageFilePath, "packageFilePath", "", "",
		"packageFilePath for chaincode")
}

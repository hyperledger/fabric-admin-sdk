/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tools

import (
	"fabric-admin-sdk/internal/configtxgen/encoder"
	"fabric-admin-sdk/internal/configtxgen/genesisconfig"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("common.tools.configtxgen")

// configtxGen
// base on Profile return block
func ConfigTxGen(config *genesisconfig.Profile, channelID string) (*cb.Block, error) {
	pgen, err := encoder.NewBootstrapper(config)
	if err != nil {
		return nil, err
	}
	genesisBlock := pgen.GenesisBlockForChannel(channelID)
	return genesisBlock, nil
}

// load profile
// file as file path
// profile_name name
func LoadProfile(configName, FABRIC_CFG_PATH string) (*genesisconfig.Profile, error) {
	return genesisconfig.Load(configName, FABRIC_CFG_PATH)
}

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package osnadmin

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	ab "github.com/hyperledger/fabric-protos-go-apiv2/orderer"
)

var seekNewest = &ab.SeekPosition{
	Type: &ab.SeekPosition_Newest{
		Newest: &ab.SeekNewest{},
	},
}

func Fetch(service ab.AtomicBroadcast_DeliverClient,
	ChannelID string, TLSCertHash []byte,
	Signer identity.SigningIdentity, BestEffort bool) (*cb.Block, error) {
	env := seekHelper(ChannelID, seekNewest, TLSCertHash, Signer, BestEffort)
	err := service.Send(env)
	if err != nil {
		return nil, err
	}
	return readBlock(service)
}

func readBlock(service ab.AtomicBroadcast_DeliverClient) (*cb.Block, error) {
	msg, err := service.Recv()
	if err != nil {
		return nil, errors.Unwrap(fmt.Errorf("error reading receiving %w", err))
	}
	switch t := msg.Type.(type) {
	case *ab.DeliverResponse_Status:
		return nil, errors.New(fmt.Sprint("can't read the block: %v", t))
	case *ab.DeliverResponse_Block:
		if resp, err := service.Recv(); err != nil { // Flush the success message
			fmt.Errorf("Failed to flush success message: %s", err)
		} else if status := resp.GetStatus(); status != cb.Status_SUCCESS {
			fmt.Errorf("Expect status to be SUCCESS, got: %s", status)
		}
		return t.Block, nil
	default:
		return nil, errors.New(fmt.Sprint("response error: unknown type %T", t))
	}
}

func seekHelper(
	channelID string,
	position *ab.SeekPosition,
	tlsCertHash []byte,
	signer identity.SigningIdentity,
	bestEffort bool,
) *cb.Envelope {
	seekInfo := &ab.SeekInfo{
		Start:    position,
		Stop:     position,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}

	if bestEffort {
		seekInfo.ErrorResponse = ab.SeekInfo_BEST_EFFORT
	}

	env, err := protoutil.CreateSignedEnvelopeWithTLSBinding(
		cb.HeaderType_DELIVER_SEEK_INFO,
		channelID,
		signer,
		seekInfo,
		int32(0),
		uint64(0),
		tlsCertHash,
	)
	if err != nil {
		return nil
	}

	return env
}

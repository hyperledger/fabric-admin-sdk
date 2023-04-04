package channel

import (
	"crypto/sha256"
	"crypto/tls"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	ab "github.com/hyperledger/fabric-protos-go-apiv2/orderer"
)

var Newest = &ab.SeekPosition{
	Type: &ab.SeekPosition_Newest{
		Newest: &ab.SeekNewest{},
	},
}

func GetNewstBlock(Service ab.AtomicBroadcast_DeliverClient, channelID string, Certificate tls.Certificate, signer identity.SigningIdentity, bestEffort bool) (*cb.Block, error) {
	var tlsCertHash []byte
	hasher := sha256.New()
	hasher.Write([]byte(Certificate.Certificate[0]))
	tlsCertHash = hasher.Sum(nil)
	err := seekNewest(Service, channelID, tlsCertHash, signer, bestEffort)
	if err != nil {
		return nil, err
	}
	return readBlock(Service)
}

func readBlock(Service ab.AtomicBroadcast_DeliverClient) (*cb.Block, error) {
	msg, err := Service.Recv()
	if err != nil {
		return nil, err
	}
	switch t := msg.Type.(type) {
	case *ab.DeliverResponse_Status:
		//logger.Infof("Expect block, but got status: %v", t)
		return nil, fmt.Errorf("can't read the block: %v", t)
	case *ab.DeliverResponse_Block:
		//logger.Infof("Received block: %v", t.Block.Header.Number)
		if resp, err := Service.Recv(); err != nil { // Flush the success message
			fmt.Printf("Failed to flush success message: %s", err)
		} else if status := resp.GetStatus(); status != cb.Status_SUCCESS {
			fmt.Printf("Expect status to be SUCCESS, got: %s", status)
		}

		return t.Block, nil
	default:
		return nil, fmt.Errorf("response error: unknown type %T", t)
	}
}

func seekNewest(Service ab.AtomicBroadcast_DeliverClient, channelID string, tlsCertHash []byte, signer identity.SigningIdentity, bestEffort bool) error {
	env := seekHelper(channelID, Newest, tlsCertHash, signer, bestEffort)
	return Service.Send(env)
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
		fmt.Printf("Error signing envelope:  %s", err)
		return nil
	}

	return env
}

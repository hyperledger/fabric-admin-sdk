package channel

import (
	"crypto/sha256"
	"crypto/tls"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	ab "github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"google.golang.org/protobuf/proto"
)

var newest = &ab.SeekPosition{
	Type: &ab.SeekPosition_Newest{
		Newest: &ab.SeekNewest{},
	},
}

func GetConfigBlockFromOrderer(Service ab.AtomicBroadcast_DeliverClient, channelID string, Certificate tls.Certificate, signer identity.SigningIdentity, bestEffort bool) (*cb.Block, error) {
	iBlock, err := GetNewstBlock(Service, channelID, Certificate, signer, bestEffort)
	if err != nil {
		return nil, err
	}
	lc, err := GetLastConfigIndexFromBlock(iBlock)
	if err != nil {
		return nil, err
	}
	return GetSpecifiedBlock(Service, channelID, Certificate, signer, bestEffort, lc)
}

func getSpecifiedBlock(Service ab.AtomicBroadcast_DeliverClient, channelID string, Certificate tls.Certificate, signer identity.SigningIdentity, bestEffort bool, num uint64) (*cb.Block, error) {
	seekPosition := &ab.SeekPosition{
		Type: &ab.SeekPosition_Specified{
			Specified: &ab.SeekSpecified{
				Number: num,
			},
		},
	}
	return getBlockBySeekPosition(Service, channelID, Certificate, signer, bestEffort, seekPosition)
}

func getBlockBySeekPosition(Service ab.AtomicBroadcast_DeliverClient, channelID string, Certificate tls.Certificate, signer identity.SigningIdentity, bestEffort bool, seekPosition *ab.SeekPosition) (*cb.Block, error) {
	var tlsCertHash []byte
	hasher := sha256.New()
	hasher.Write([]byte(Certificate.Certificate[0]))
	tlsCertHash = hasher.Sum(nil)
	env := seekHelper(channelID, seekPosition, tlsCertHash, signer, bestEffort)
	err := Service.Send(env)
	if err != nil {
		return nil, err
	}
	return readBlock(Service)
}

func GetNewstBlock(Service ab.AtomicBroadcast_DeliverClient, channelID string, Certificate tls.Certificate, signer identity.SigningIdentity, bestEffort bool) (*cb.Block, error) {
	return getBlockBySeekPosition(Service, channelID, Certificate, signer, bestEffort, newest)
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

// GetLastConfigIndexFromBlock retrieves the index of the last config block as
// encoded in the block metadata
func GetLastConfigIndexFromBlock(block *cb.Block) (uint64, error) {
	m, err := GetMetadataFromBlock(block, cb.BlockMetadataIndex_SIGNATURES)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve metadata %w", err)
	}
	// TODO FAB-15864 Remove this fallback when we can stop supporting upgrade from pre-1.4.1 orderer
	if len(m.Value) == 0 {
		// TODO cb.BlockMetadataIndex_LAST_CONFIG
		m, err := GetMetadataFromBlock(block, 1)
		if err != nil {
			return 0, fmt.Errorf("failed to retrieve metadata %w", err)
		}
		lc := &cb.LastConfig{}
		err = proto.Unmarshal(m.Value, lc)
		if err != nil {
			return 0, fmt.Errorf("error unmarshalling LastConfig %w", err)
		}
		return lc.Index, nil
	}

	obm := &cb.OrdererBlockMetadata{}
	err = proto.Unmarshal(m.Value, obm)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal orderer block metadata %w", err)
	}
	return obm.LastConfig.Index, nil
}

// GetMetadataFromBlock retrieves metadata at the specified index.
func GetMetadataFromBlock(block *cb.Block, index cb.BlockMetadataIndex) (*cb.Metadata, error) {
	if block.Metadata == nil {
		return nil, fmt.Errorf("no metadata in block")
	}

	if len(block.Metadata.Metadata) <= int(index) {
		return nil, fmt.Errorf("no metadata at index [%s]", index)
	}

	md := &cb.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[index], md)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata at index [%s] %w", index, err)
	}
	return md, nil
}

package channel

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-admin-sdk/internal/protoutil"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
	ab "github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var newest = &ab.SeekPosition{
	Type: &ab.SeekPosition_Newest{
		Newest: &ab.SeekNewest{},
	},
}

func GetConfigBlockFromOrderer(ctx context.Context, connection grpc.ClientConnInterface, id identity.SigningIdentity, channelID string, certificate tls.Certificate) (*cb.Block, error) {
	abClient := ab.NewAtomicBroadcastClient(connection)
	deliverClient, err := abClient.Deliver(ctx)
	if err != nil {
		return nil, err
	}
	iBlock, err := getNewestBlock(deliverClient, channelID, certificate, id, true)
	if err != nil {
		return nil, err
	}
	lc, err := getLastConfigIndexFromBlock(iBlock)
	if err != nil {
		return nil, err
	}
	return getSpecifiedBlock(deliverClient, channelID, certificate, id, true, lc)
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

func getNewestBlock(Service ab.AtomicBroadcast_DeliverClient, channelID string, Certificate tls.Certificate, signer identity.SigningIdentity, bestEffort bool) (*cb.Block, error) {
	return getBlockBySeekPosition(Service, channelID, Certificate, signer, bestEffort, newest)
}

func readBlock(Service ab.AtomicBroadcast_DeliverClient) (*cb.Block, error) {
	msg, err := Service.Recv()
	if err != nil {
		return nil, err
	}
	switch t := msg.GetType().(type) {
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
func getLastConfigIndexFromBlock(block *cb.Block) (uint64, error) {
	m, err := getMetadataFromBlock(block, cb.BlockMetadataIndex_SIGNATURES)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve metadata %w", err)
	}
	// TODO FAB-15864 Remove this fallback when we can stop supporting upgrade from pre-1.4.1 orderer
	if len(m.GetValue()) == 0 {
		// TODO cb.BlockMetadataIndex_LAST_CONFIG
		m, err := getMetadataFromBlock(block, 1)
		if err != nil {
			return 0, fmt.Errorf("failed to retrieve metadata %w", err)
		}
		lc := &cb.LastConfig{}
		err = proto.Unmarshal(m.GetValue(), lc)
		if err != nil {
			return 0, fmt.Errorf("error unmarshalling LastConfig %w", err)
		}
		return lc.GetIndex(), nil
	}

	obm := &cb.OrdererBlockMetadata{}
	err = proto.Unmarshal(m.GetValue(), obm)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal orderer block metadata %w", err)
	}
	return obm.GetLastConfig().GetIndex(), nil
}

// GetMetadataFromBlock retrieves metadata at the specified index.
func getMetadataFromBlock(block *cb.Block, index cb.BlockMetadataIndex) (*cb.Metadata, error) {
	if block.GetMetadata() == nil {
		return nil, errors.New("no metadata in block")
	}

	if len(block.GetMetadata().GetMetadata()) <= int(index) {
		return nil, fmt.Errorf("no metadata at index [%s]", index)
	}

	md := &cb.Metadata{}
	err := proto.Unmarshal(block.GetMetadata().GetMetadata()[index], md)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata at index [%s] %w", index, err)
	}
	return md, nil
}

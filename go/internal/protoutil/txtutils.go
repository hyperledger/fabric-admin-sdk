package protoutil

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fabric-admin-sdk/internal/pkg/identity"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

// GetSignedProposal returns a signed proposal given a Proposal message and a
// signing identity
func GetSignedProposal(prop *peer.Proposal, signer *identity.CryptoImpl) (*peer.SignedProposal, error) {
	// check for nil argument
	if prop == nil || signer == nil {
		return nil, errors.New("nil arguments")
	}

	propBytes, err := proto.Marshal(prop)
	if err != nil {
		return nil, err
	}

	signature, err := signer.Sign(propBytes)
	if err != nil {
		return nil, err
	}

	return &peer.SignedProposal{ProposalBytes: propBytes, Signature: signature}, nil
}

// CreateProposalFromCIS returns a proposal given a serialized identity and a
// ChaincodeInvocationSpec
func CreateProposalFromCIS(typ common.HeaderType, channelID string, cis *peer.ChaincodeInvocationSpec, creator []byte) (*peer.Proposal, string, error) {
	return CreateChaincodeProposal(typ, channelID, cis, creator)
}

// CreateChaincodeProposal creates a proposal from given input.
// It returns the proposal and the transaction id associated to the proposal
func CreateChaincodeProposal(typ common.HeaderType, channelID string, cis *peer.ChaincodeInvocationSpec, creator []byte) (*peer.Proposal, string, error) {
	return CreateChaincodeProposalWithTransient(typ, channelID, cis, creator, nil)
}

// CreateChaincodeProposalWithTransient creates a proposal from given input
// It returns the proposal and the transaction id associated to the proposal
func CreateChaincodeProposalWithTransient(typ common.HeaderType, channelID string, cis *peer.ChaincodeInvocationSpec, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	// generate a random nonce
	nonce, err := getRandomNonce()
	if err != nil {
		return nil, "", err
	}

	// compute txid
	txid := ComputeTxID(nonce, creator)

	return CreateChaincodeProposalWithTxIDNonceAndTransient(txid, typ, channelID, cis, nonce, creator, transientMap)
}

// ComputeTxID computes TxID as the Hash computed
// over the concatenation of nonce and creator.
func ComputeTxID(nonce, creator []byte) string {
	// TODO: Get the Hash function to be used from
	// channel configuration
	hasher := sha256.New()
	hasher.Write(nonce)
	hasher.Write(creator)
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetRandomNonce() ([]byte, error) {
	return getRandomNonce()
}

func getRandomNonce() ([]byte, error) {
	key := make([]byte, 24)

	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("error getting random bytes %w", err)
	}
	return key, nil
}

// CreateChaincodeProposalWithTxIDNonceAndTransient creates a proposal from
// given input
func CreateChaincodeProposalWithTxIDNonceAndTransient(txid string, typ common.HeaderType, channelID string, cis *peer.ChaincodeInvocationSpec, nonce, creator []byte, transientMap map[string][]byte) (*peer.Proposal, string, error) {
	ccHdrExt := &peer.ChaincodeHeaderExtension{ChaincodeId: cis.ChaincodeSpec.ChaincodeId}
	ccHdrExtBytes, err := proto.Marshal(ccHdrExt)
	if err != nil {
		return nil, "", fmt.Errorf("error marshaling ChaincodeHeaderExtension %w", err)
	}

	cisBytes, err := proto.Marshal(cis)
	if err != nil {
		return nil, "", fmt.Errorf("error marshaling ChaincodeInvocationSpec %w", err)
	}

	ccPropPayload := &peer.ChaincodeProposalPayload{Input: cisBytes, TransientMap: transientMap}
	ccPropPayloadBytes, err := proto.Marshal(ccPropPayload)
	if err != nil {
		return nil, "", errors.Unwrap(fmt.Errorf("error marshaling ChaincodeProposalPayload %w", err))
	}

	// TODO: epoch is now set to zero. This must be changed once we
	// get a more appropriate mechanism to handle it in.
	var epoch uint64

	timestamp, err := ptypes.TimestampProto(time.Now().UTC())
	if err != nil {
		return nil, "", fmt.Errorf("error validating Timestamp %w", err)
	}

	hdr := &common.Header{
		ChannelHeader: MarshalOrPanic(
			&common.ChannelHeader{
				Type:      int32(typ),
				TxId:      txid,
				Timestamp: timestamp,
				ChannelId: channelID,
				Extension: ccHdrExtBytes,
				Epoch:     epoch,
			},
		),
		SignatureHeader: MarshalOrPanic(
			&common.SignatureHeader{
				Nonce:   nonce,
				Creator: creator,
			},
		),
	}

	hdrBytes, err := proto.Marshal(hdr)
	if err != nil {
		return nil, "", err
	}

	prop := &peer.Proposal{
		Header:  hdrBytes,
		Payload: ccPropPayloadBytes,
	}
	return prop, txid, nil
}

// MarshalOrPanic serializes a protobuf message and panics if this
// operation fails
func MarshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

// CreateSignedEnvelope creates a signed envelope of the desired type, with
// marshaled dataMsg and signs it
func CreateSignedEnvelope(
	txType common.HeaderType,
	channelID string,
	signer identity.SignerSerializer,
	dataMsg proto.Message,
	msgVersion int32,
	epoch uint64,
) (*common.Envelope, error) {
	return CreateSignedEnvelopeWithTLSBinding(txType, channelID, signer, dataMsg, msgVersion, epoch, nil)
}

// CreateSignedEnvelopeWithTLSBinding creates a signed envelope of the desired
// type, with marshaled dataMsg and signs it. It also includes a TLS cert hash
// into the channel header
func CreateSignedEnvelopeWithTLSBinding(
	txType common.HeaderType,
	channelID string,
	signer identity.SignerSerializer,
	dataMsg proto.Message,
	msgVersion int32,
	epoch uint64,
	tlsCertHash []byte,
) (*common.Envelope, error) {
	payloadChannelHeader := MakeChannelHeader(txType, msgVersion, channelID, epoch)
	payloadChannelHeader.TlsCertHash = tlsCertHash
	var err error
	payloadSignatureHeader := &common.SignatureHeader{}

	if signer != nil {
		payloadSignatureHeader, err = NewSignatureHeader(signer)
		if err != nil {
			return nil, err
		}
	}

	data, err := proto.Marshal(dataMsg)
	if err != nil {
		return nil, fmt.Errorf("error marshaling %w", err)
	}

	paylBytes := MarshalOrPanic(
		&common.Payload{
			Header: MakePayloadHeader(payloadChannelHeader, payloadSignatureHeader),
			Data:   data,
		},
	)

	var sig []byte
	if signer != nil {
		sig, err = signer.Sign(paylBytes)
		if err != nil {
			return nil, err
		}
	}

	env := &common.Envelope{
		Payload:   paylBytes,
		Signature: sig,
	}

	return env, nil
}

// MakeChannelHeader creates a ChannelHeader.
func MakeChannelHeader(headerType common.HeaderType, version int32, chainID string, epoch uint64) *common.ChannelHeader {
	return &common.ChannelHeader{
		Type:    int32(headerType),
		Version: version,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
			Nanos:   0,
		},
		ChannelId: chainID,
		Epoch:     epoch,
	}
}

// NewSignatureHeader returns a SignatureHeader with a valid nonce.
func NewSignatureHeader(id identity.SignerSerializer) (*common.SignatureHeader, error) {
	creator, err := id.Serialize()
	if err != nil {
		return nil, err
	}
	nonce, err := CreateNonce()
	if err != nil {
		return nil, err
	}

	return &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}, nil
}

func BlockDataHash(b *common.BlockData) []byte {
	sum := sha256.Sum256(bytes.Join(b.Data, nil))
	return sum[:]
}

// NewBlock constructs a block with no data and no metadata.
func NewBlock(seqNum uint64, previousHash []byte) *common.Block {
	block := &common.Block{}
	block.Header = &common.BlockHeader{}
	block.Header.Number = seqNum
	block.Header.PreviousHash = previousHash
	block.Header.DataHash = []byte{}
	block.Data = &common.BlockData{}

	var metadataContents [][]byte
	for i := 0; i < len(common.BlockMetadataIndex_name); i++ {
		metadataContents = append(metadataContents, []byte{})
	}
	block.Metadata = &common.BlockMetadata{Metadata: metadataContents}

	return block
}

// MakeSignatureHeader creates a SignatureHeader.
func MakeSignatureHeader(serializedCreatorCertChain []byte, nonce []byte) *common.SignatureHeader {
	return &common.SignatureHeader{
		Creator: serializedCreatorCertChain,
		Nonce:   nonce,
	}
}

// SetTxID generates a transaction id based on the provided signature header
// and sets the TxId field in the channel header
func SetTxID(channelHeader *common.ChannelHeader, signatureHeader *common.SignatureHeader) {
	channelHeader.TxId = ComputeTxID(
		signatureHeader.Nonce,
		signatureHeader.Creator,
	)
}

// CreateNonceOrPanic generates a nonce using the common/crypto package
// and panics if this operation fails.
func CreateNonceOrPanic() []byte {
	nonce, err := CreateNonce()
	if err != nil {
		panic(err)
	}
	return nonce
}

// CreateNonce generates a nonce using the common/crypto package.
func CreateNonce() ([]byte, error) {
	nonce, err := getRandomNonce()
	return nonce, errors.Unwrap(fmt.Errorf("error generating random nonce %w", err))
}

// MakePayloadHeader creates a Payload Header.
func MakePayloadHeader(ch *common.ChannelHeader, sh *common.SignatureHeader) *common.Header {
	return &common.Header{
		ChannelHeader:   MarshalOrPanic(ch),
		SignatureHeader: MarshalOrPanic(sh),
	}
}

// UnmarshalHeader unmarshals bytes to a Header
func UnmarshalHeader(bytes []byte) (*common.Header, error) {
	hdr := &common.Header{}
	err := proto.Unmarshal(bytes, hdr)
	return hdr, errors.Unwrap(fmt.Errorf("error unmarshaling Header %w", err))
}

// UnmarshalChaincodeProposalPayload unmarshals bytes to a ChaincodeProposalPayload
func UnmarshalChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, errors.Unwrap(fmt.Errorf("error unmarshaling ChaincodeProposalPayload %w", err))
}

// UnmarshalSignatureHeader unmarshals bytes to a SignatureHeader
func UnmarshalSignatureHeader(bytes []byte) (*common.SignatureHeader, error) {
	sh := &common.SignatureHeader{}
	err := proto.Unmarshal(bytes, sh)
	return sh, errors.Unwrap(fmt.Errorf("error unmarshaling SignatureHeader %w", err))
}

// GetBytesProposalPayloadForTx takes a ChaincodeProposalPayload and returns
// its serialized version according to the visibility field
func GetBytesProposalPayloadForTx(
	payload *peer.ChaincodeProposalPayload,
) ([]byte, error) {
	// check for nil argument
	if payload == nil {
		return nil, errors.New("nil arguments")
	}

	// strip the transient bytes off the payload
	cppNoTransient := &peer.ChaincodeProposalPayload{Input: payload.Input, TransientMap: nil}
	cppBytes, err := GetBytesChaincodeProposalPayload(cppNoTransient)
	if err != nil {
		return nil, err
	}

	return cppBytes, nil
}

// GetBytesChaincodeProposalPayload gets the chaincode proposal payload
func GetBytesChaincodeProposalPayload(cpp *peer.ChaincodeProposalPayload) ([]byte, error) {
	cppBytes, err := proto.Marshal(cpp)
	return cppBytes, errors.Unwrap(fmt.Errorf("error marshaling ChaincodeProposalPayload %w", err))
}

// GetBytesChaincodeActionPayload get the bytes of ChaincodeActionPayload from
// the message
func GetBytesChaincodeActionPayload(cap *peer.ChaincodeActionPayload) ([]byte, error) {
	capBytes, err := proto.Marshal(cap)
	return capBytes, errors.Unwrap(fmt.Errorf("error marshaling ChaincodeActionPayload %w", err))
}

// GetBytesTransaction get the bytes of Transaction from the message
func GetBytesTransaction(tx *peer.Transaction) ([]byte, error) {
	bytes, err := proto.Marshal(tx)
	return bytes, errors.Unwrap(fmt.Errorf("error unmarshaling Transaction %w", err))
}

// GetBytesPayload get the bytes of Payload from the message
func GetBytesPayload(payl *common.Payload) ([]byte, error) {
	bytes, err := proto.Marshal(payl)
	return bytes, errors.Unwrap(fmt.Errorf("error marshaling Payload %w", err))
}

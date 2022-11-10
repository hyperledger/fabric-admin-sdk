package channel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fabric-admin-sdk/internal/osnadmin"
	"fabric-admin-sdk/internal/pkg/identity"
	"fabric-admin-sdk/resource"
	"fmt"
	"io/ioutil"
	"net/http"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

func CreateChannel(osnURL string, block *cb.Block, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (*http.Response, error) {
	block_byte := protoutil.MarshalOrPanic(block)
	return osnadmin.Join(osnURL, block_byte, caCertPool, tlsClientCert)
}

func JoinChannel(block *cb.Block, Signer identity.CryptoImpl, connection pb.EndorserClient) error {
	spec, err := getJoinCCSpec(block)
	if err != nil {
		return err
	}
	return executeJoin(Signer, connection, spec)
}

const (
	UndefinedParamValue = ""
)

func getJoinCCSpec(block *cb.Block) (*pb.ChaincodeSpec, error) {
	block_byte := protoutil.MarshalOrPanic(block)
	// Build the spec
	input := &pb.ChaincodeInput{Args: [][]byte{[]byte("JoinChain"), block_byte}}

	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
		ChaincodeId: &pb.ChaincodeID{Name: "cscc"},
		Input:       input,
	}

	return spec, nil
}

func executeJoin(Signer identity.CryptoImpl, endorsementClinet pb.EndorserClient, spec *pb.ChaincodeSpec) (err error) {
	// Build the ChaincodeInvocationSpec message
	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	creator, err := Signer.Serialize()
	if err != nil {
		return fmt.Errorf("Error serializing identity for reason %s", err)
	}

	var prop *pb.Proposal
	prop, _, err = protoutil.CreateProposalFromCIS(cb.HeaderType_CONFIG, "", invocation, creator)
	if err != nil {
		return fmt.Errorf("Error creating proposal for join %s", err)
	}

	var signedProp *pb.SignedProposal

	signedProp, err = protoutil.GetSignedProposal(prop, Signer)
	if err != nil {
		return fmt.Errorf("Error creating signed proposal %s", err)
	}

	var proposalResp *pb.ProposalResponse
	proposalResp, err = endorsementClinet.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return err
	}

	if proposalResp == nil {
		return errors.New("nil proposal response")
	}

	if proposalResp.Response.Status != 0 && proposalResp.Response.Status != 200 {
		return errors.New(fmt.Sprintf("bad proposal response %d: %s", proposalResp.Response.Status, proposalResp.Response.Message))
	}
	return nil
}

func ListChannel(osnURL string, caCertPool *x509.CertPool, tlsClientCert tls.Certificate) (resource.ChannelList, error) {

	var channels resource.ChannelList
	resp, err := osnadmin.ListAllChannels(osnURL, caCertPool, tlsClientCert)
	if err != nil {
		return channels, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return channels, err
	}

	if err := json.Unmarshal(body, &channels); err != nil {
		return channels, err
	}

	return channels, nil
}

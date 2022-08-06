package chaincode

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity"
	"fmt"
	"io/ioutil"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/sw"
	"github.com/hyperledger/fabric/core/common/ccpackage"
	"github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/pkg/errors"
)

const (
	chainFuncName = "chaincode"
)

func GetCCPackage(buf []byte, bccsp bccsp.BCCSP) (ccprovider.CCPackage, error) {
	// try raw CDS
	cds := &ccprovider.CDSPackage{GetHasher: bccsp}
	if ccdata, err := cds.InitFromBuffer(buf); err != nil {
		fmt.Println("cds init err", err)
		cds = nil
	} else {
		err = cds.ValidateCC(ccdata)
		if err != nil {
			fmt.Println("cds validate err", err)
			cds = nil
		}
	}

	// try signed CDS
	scds := &ccprovider.SignedCDSPackage{GetHasher: bccsp}
	if ccdata, err := scds.InitFromBuffer(buf); err != nil {
		fmt.Println("scds init err", err)
		scds = nil
	} else {
		err = scds.ValidateCC(ccdata)
		if err != nil {
			fmt.Println("scds validate err", err)
			scds = nil
		}
	}

	if cds != nil && scds != nil {
		// Both were unmarshaled successfully, this is exactly why the approach of
		// hoping proto fails for bad inputs is fatally flawed.
		//ccproviderLogger.Errorf("Could not determine chaincode package type, guessing SignedCDS")
		return scds, nil
	}

	if cds != nil {
		return cds, nil
	}

	if scds != nil {
		return scds, nil
	}

	return nil, errors.New("could not unmarshal chaincode package to CDS or SignedCDS")
}

func InstallChainCode(PackageFile, ChaincodeName, ChaincodeVersion string, Signer identity.CryptoImpl, endorsementClinet pb.EndorserClient) error {
	var proposal *pb.Proposal
	var signedProposal *pb.SignedProposal

	ccPkgBytes, err := ioutil.ReadFile(PackageFile)
	if err != nil {
		return err
	}
	csp, err := sw.NewWithParams(256, "SHA2", sw.NewDummyKeyStore())
	if err != nil {
		return err
	}

	ccpack, err := GetCCPackage(ccPkgBytes, csp)
	if err != nil {
		return err
	}

	o := ccpack.GetPackageObject()
	cds, ok := o.(*pb.ChaincodeDeploymentSpec)
	if !ok || cds == nil {
		// try Envelope next
		env, ok := o.(*cb.Envelope)
		if !ok || env == nil {
			return errors.New("error extracting valid chaincode package")
		}

		// this will check for a valid package Envelope
		_, sCDS, err := ccpackage.ExtractSignedCCDepSpec(env)
		if err != nil {
			return errors.WithMessage(err, "error extracting valid signed chaincode package")
		}

		// ...and get the CDS at last
		cds, err = protoutil.UnmarshalChaincodeDeploymentSpec(sCDS.ChaincodeDeploymentSpec)
		if err != nil {
			return errors.WithMessage(err, "error extracting chaincode deployment spec")
		}
	}

	creator, err := Signer.Serialize()
	if err != nil {
		return errors.WithMessage(err, "error serializing identity")
	}

	proposal, _, err = protoutil.CreateInstallProposalFromCDS(cds, creator)
	if err != nil {
		return errors.WithMessagef(err, "error creating proposal for %s", chainFuncName)
	}

	signedProposal, err = protoutil.GetSignedProposal(proposal, Signer)
	if err != nil {
		return errors.WithMessagef(err, "error creating signed proposal for %s", chainFuncName)
	}
	// Submit Install Proposal
	proposalResponse, err := endorsementClinet.ProcessProposal(context.Background(), signedProposal)
	if err != nil {
		return errors.WithMessage(err, "error during install:"+err.Error())
	}

	if proposalResponse == nil {
		return errors.New("error during install: received nil proposal response")
	}

	if proposalResponse.Response == nil {
		return errors.New("error during install: received proposal response with nil response")
	}

	if proposalResponse.Response.Status != int32(cb.Status_SUCCESS) {
		return errors.Errorf("install failed with status: %d - %s", proposalResponse.Response.Status, proposalResponse.Response.Message)
	}

	return nil
}

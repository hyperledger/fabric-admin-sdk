package chaincode

import (
	"context"
	"fabric-admin-sdk/internal/pkg/identity"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/core/common/ccpackage"
	"github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

const (
	chainFuncName = "chaincode"
)

func InstallChainCode(Path, PackageFile, ChaincodeName, ChaincodeVersion string, Signer identity.CryptoImpl, endorsementClinet pb.EndorserClient) error {
	//csp, err := sw.NewWithParams(256, "SHA2", sw.NewDummyKeyStore())
	//if err != nil {
	//	return err
	//}
	/*ccPkgMsg, err := getChaincodePackageMessage(PackageFile, Path, ChaincodeVersion, ChaincodeName, csp)
	if err != nil {
		return err
	}*/
	pkgBytes, err := ioutil.ReadFile(PackageFile)
	if err != nil {
		return errors.WithMessagef(err, "failed to read chaincode package at '%s'", PackageFile)
	}

	serializedSigner, err := Signer.Serialize()
	if err != nil {
		return errors.Wrap(err, "failed to serialize signer")
	}

	proposal, err := createInstallProposal(pkgBytes, serializedSigner)
	if err != nil {
		return err
	}

	signedProposal, err := protoutil.GetSignedProposal(proposal, Signer)
	if err != nil {
		return errors.WithMessagef(err, "error creating signed proposal for %s", chainFuncName)
	}

	return submitInstallProposal(endorsementClinet, signedProposal)
}

func getChaincodePackageMessage(PackageFile, Path, Version, Name string, CryptoProvider bccsp.BCCSP) (proto.Message, error) {
	// if no package provided, create one
	if PackageFile == "" {
		if Path == "" || Version == "" || Name == "" {
			return nil, errors.Errorf("must supply value for %s name, path and version parameters", chainFuncName)
		}
		// generate a raw ChaincodeDeploymentSpec
		ccPkgMsg, err := genChaincodeDeploymentSpec(Path, Name, Version)
		if err != nil {
			return nil, err
		}
		return ccPkgMsg, nil
	}

	// read in a package generated by the "package" sub-command (and perhaps signed
	// by multiple owners with the "signpackage" sub-command)
	// var cds *pb.ChaincodeDeploymentSpec
	ccPkgMsg, cds, err := getPackageFromFile(PackageFile, CryptoProvider)
	if err != nil {
		return nil, err
	}

	// get the chaincode details from cds
	cName := cds.ChaincodeSpec.ChaincodeId.Name
	cVersion := cds.ChaincodeSpec.ChaincodeId.Version

	// if user provided chaincodeName, use it for validation
	if Name != "" && Name != cName {
		return nil, errors.Errorf("chaincode name %s does not match name %s in package", Name, cName)
	}

	// if user provided chaincodeVersion, use it for validation
	if Version != "" && Version != cVersion {
		return nil, errors.Errorf("chaincode version %s does not match version %s in packages", Version, cVersion)
	}

	return ccPkgMsg, nil
}

/*func createInstallProposal(Signer identity.CryptoImpl, msg proto.Message) (*pb.Proposal, error) {
	creator, err := Signer.Serialize()
	if err != nil {
		return nil, errors.WithMessage(err, "error serializing identity")
	}

	prop, _, err := protoutil.CreateInstallProposalFromCDS(msg, creator)
	if err != nil {
		return nil, errors.WithMessagef(err, "error creating proposal for %s", chainFuncName)
	}

	return prop, nil
}*/

func submitInstallProposal(endorsementClinet pb.EndorserClient, signedProposal *pb.SignedProposal) error {
	// install is currently only supported for one peer
	proposalResponse, err := endorsementClinet.ProcessProposal(context.Background(), signedProposal)
	if err != nil {
		return errors.WithMessage(err, "error endorsing chaincode install")
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
	fmt.Println("Installed remotely: ", proposalResponse)

	return nil
}

func genChaincodeDeploymentSpec(chaincodepath, chaincodeName, chaincodeVersion string) (*pb.ChaincodeDeploymentSpec, error) {
	if existed, _ := ccprovider.ChaincodePackageExists(chaincodeName, chaincodeVersion); existed {
		return nil, errors.Errorf("chaincode %s:%s already exists", chaincodeName, chaincodeVersion)
	}

	spec, err := getChaincodeSpec(chaincodepath, chaincodeName, chaincodeVersion)
	if err != nil {
		return nil, err
	}

	cds, err := getChaincodeDeploymentSpec(spec, true)
	if err != nil {
		return nil, errors.WithMessagef(err, "error getting chaincode deployment spec for %s", chaincodeName)
	}

	return cds, nil
}

// getPackageFromFile get the chaincode package from file and the extracted ChaincodeDeploymentSpec
func getPackageFromFile(ccPkgFile string, cryptoProvider bccsp.BCCSP) (proto.Message, *pb.ChaincodeDeploymentSpec, error) {
	ccPkgBytes, err := ioutil.ReadFile(ccPkgFile)
	if err != nil {
		return nil, nil, err
	}

	// the bytes should be a valid package (CDS or SignedCDS)
	ccpack, err := GetCCPackage(ccPkgBytes, cryptoProvider)
	if err != nil {
		return nil, nil, errors.New(err.Error() + " " + strconv.Itoa(len(ccPkgBytes)) + " " + ccPkgFile)
	}

	// either CDS or Envelope
	o := ccpack.GetPackageObject()

	// try CDS first
	cds, ok := o.(*pb.ChaincodeDeploymentSpec)
	if !ok || cds == nil {
		// try Envelope next
		env, ok := o.(*cb.Envelope)
		if !ok || env == nil {
			return nil, nil, errors.New("error extracting valid chaincode package")
		}

		// this will check for a valid package Envelope
		_, sCDS, err := ccpackage.ExtractSignedCCDepSpec(env)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "error extracting valid signed chaincode package")
		}

		// ...and get the CDS at last
		cds, err = protoutil.UnmarshalChaincodeDeploymentSpec(sCDS.ChaincodeDeploymentSpec)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "error extracting chaincode deployment spec")
		}

		/*err = platformRegistry.ValidateDeploymentSpec(cds.ChaincodeSpec.Type.String(), cds.CodePackage)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "chaincode deployment spec validation failed")
		}*/
	}

	return o, cds, nil
}

// getChaincodeDeploymentSpec get chaincode deployment spec given the chaincode spec
func getChaincodeDeploymentSpec(spec *pb.ChaincodeSpec, crtPkg bool) (*pb.ChaincodeDeploymentSpec, error) {
	var codePackageBytes []byte
	if crtPkg {
		var err error
		if err = checkSpec(spec); err != nil {
			return nil, err
		}

		/*codePackageBytes, err = platformRegistry.GetDeploymentPayload(spec.Type.String(), spec.ChaincodeId.Path)
		if err != nil {
			return nil, errors.WithMessage(err, "error getting chaincode package bytes")
		}
		chaincodePath, err := platformRegistry.NormalizePath(spec.Type.String(), spec.ChaincodeId.Path)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to normalize chaincode path")
		}
		spec.ChaincodeId.Path = chaincodePath*/
	}

	return &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}, nil
}

// checkSpec to see if chaincode resides within current package capture for language.
func checkSpec(spec *pb.ChaincodeSpec) error {
	// Don't allow nil value
	if spec == nil {
		return errors.New("expected chaincode specification, nil received")
	}
	if spec.ChaincodeId == nil {
		return errors.New("expected chaincode ID, nil received")
	}

	return nil //platformRegistry.ValidateSpec(spec.Type.String(), spec.ChaincodeId.Path)
}

func getChaincodeSpec(chaincodePath, chaincodeName, chaincodeVersion string) (*pb.ChaincodeSpec, error) {
	spec := &pb.ChaincodeSpec{}
	/*if err := checkChaincodeCmdParams(cmd); err != nil {
		// unset usage silence because it's a command line usage error
		cmd.SilenceUsage = false
		return spec, err
	}*/

	// Build the spec
	/*input := chaincodeInput{}
	if err := json.Unmarshal([]byte(chaincodeCtorJSON), &input); err != nil {
		return spec, errors.Wrap(err, "chaincode argument error")
	}
	input.IsInit = isInit*/

	//chaincodeLang = strings.ToUpper(chaincodeLang)
	spec = &pb.ChaincodeSpec{
		//Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[chaincodeLang]),
		ChaincodeId: &pb.ChaincodeID{Path: chaincodePath, Name: chaincodeName, Version: chaincodeVersion},
		//Input:       &input.ChaincodeInput,
	}
	return spec, nil
}

func GetCCPackage(buf []byte, bccsp bccsp.BCCSP) (ccprovider.CCPackage, error) {
	// try raw CDS
	var cds_err error
	var scds_err error
	cds := &ccprovider.CDSPackage{GetHasher: bccsp}
	if ccdata, err := cds.InitFromBuffer(buf); err != nil {
		cds = nil
		cds_err = err
	} else {
		err = cds.ValidateCC(ccdata)
		if err != nil {
			cds = nil
			cds_err = err
		}
	}

	// try signed CDS
	scds := &ccprovider.SignedCDSPackage{GetHasher: bccsp}
	if ccdata, err := scds.InitFromBuffer(buf); err != nil {
		scds = nil
		scds_err = err
	} else {
		err = scds.ValidateCC(ccdata)
		if err != nil {
			scds = nil
			scds_err = err
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

	return nil, errors.New("could not unmarshal chaincode package to CDS [" + cds_err.Error() + "]or SignedCDS [" + scds_err.Error() + "]")
}

func createInstallProposal(pkgBytes []byte, creatorBytes []byte) (*pb.Proposal, error) {
	installChaincodeArgs := &lb.InstallChaincodeArgs{
		ChaincodeInstallPackage: pkgBytes,
	}

	installChaincodeArgsBytes, err := proto.Marshal(installChaincodeArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal InstallChaincodeArgs")
	}

	ccInput := &pb.ChaincodeInput{Args: [][]byte{[]byte("InstallChaincode"), installChaincodeArgsBytes}}

	cis := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: &pb.ChaincodeSpec{
			ChaincodeId: &pb.ChaincodeID{Name: "_lifecycle"},
			Input:       ccInput,
		},
	}

	proposal, _, err := protoutil.CreateProposalFromCIS(cb.HeaderType_ENDORSER_TRANSACTION, "", cis, creatorBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create proposal for ChaincodeInvocationSpec")
	}

	return proposal, nil
}

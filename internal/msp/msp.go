package msp

import (
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// ProviderType indicates the type of an identity provider
type ProviderType int

const (
	FABRIC ProviderType = iota // MSP is of FABRIC type
	IDEMIX                     // MSP is of IDEMIX type
	OTHER                      // MSP is of OTHER TYPE

	// NOTE: as new types are added to this set,
	// the mspTypes map below must be extended
)

var mspTypeStrings = map[ProviderType]string{
	FABRIC: "bccsp",
	IDEMIX: "idemix",
}

const (
	cacerts              = "cacerts"
	admincerts           = "admincerts"
	signcerts            = "signcerts"
	keystore             = "keystore"
	intermediatecerts    = "intermediatecerts"
	crlsfolder           = "crls"
	configfilename       = "config.yaml"
	tlscacerts           = "tlscacerts"
	tlsintermediatecerts = "tlsintermediatecerts"
)

// GetVerifyingMspConfig returns an MSP config given directory, ID and type
func GetVerifyingMspConfig(dir, ID, mspType string) (*msp.MSPConfig, error) {
	switch mspType {
	case ProviderTypeToString(FABRIC):
		return getMspConfig(dir, ID, nil)
	case ProviderTypeToString(IDEMIX):
		return GetIdemixMspConfig(dir, ID)
	default:
		return nil, fmt.Errorf("unknown MSP type '%s'", mspType)
	}
}

// ProviderTypeToString returns a string that represents the ProviderType integer
func ProviderTypeToString(id ProviderType) string {
	if res, found := mspTypeStrings[id]; found {
		return res
	}

	return ""
}

//nolint:cyclop,gocognit
func getMspConfig(dir string, ID string, sigid *msp.SigningIdentityInfo) (*msp.MSPConfig, error) {
	cacertDir := filepath.Join(dir, cacerts)
	admincertDir := filepath.Join(dir, admincerts)
	intermediatecertsDir := filepath.Join(dir, intermediatecerts)
	crlsDir := filepath.Join(dir, crlsfolder)
	configFile := filepath.Join(dir, configfilename)
	tlscacertDir := filepath.Join(dir, tlscacerts)
	tlsintermediatecertsDir := filepath.Join(dir, tlsintermediatecerts)

	cacerts, err := getPemMaterialFromDir(cacertDir)
	if err != nil || len(cacerts) == 0 {
		return nil, fmt.Errorf("could not load a valid ca certificate from directory %s %w", cacertDir, err)
	}

	admincert, err := getPemMaterialFromDir(admincertDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("could not load a valid admin certificate from directory %s %w", admincertDir, err)
	}

	intermediatecerts, err := getPemMaterialFromDir(intermediatecertsDir)
	if os.IsNotExist(err) {
		fmt.Printf("Intermediate certs folder not found at [%s]. Skipping. [%s]", intermediatecertsDir, err)
	} else if err != nil {
		return nil, fmt.Errorf("failed loading intermediate ca certs at [%s] %w", intermediatecertsDir, err)
	}

	tlsCACerts, err := getPemMaterialFromDir(tlscacertDir)
	tlsIntermediateCerts := [][]byte{}
	if os.IsNotExist(err) {
		fmt.Printf("TLS CA certs folder not found at [%s]. Skipping and ignoring TLS intermediate CA folder. [%s]", tlsintermediatecertsDir, err)
	} else if err != nil {
		return nil, fmt.Errorf("failed loading TLS ca certs at [%s] %w", tlsintermediatecertsDir, err)
	} else if len(tlsCACerts) != 0 {
		tlsIntermediateCerts, err = getPemMaterialFromDir(tlsintermediatecertsDir)
		if os.IsNotExist(err) {
			fmt.Printf("TLS intermediate certs folder not found at [%s]. Skipping. [%s]", tlsintermediatecertsDir, err)
		} else if err != nil {
			return nil, fmt.Errorf("failed loading TLS intermediate ca certs at [%s] %w", tlsintermediatecertsDir, err)
		}
	} else {
		fmt.Printf("TLS CA certs folder at [%s] is empty. Skipping.", tlsintermediatecertsDir)
	}

	crls, err := getPemMaterialFromDir(crlsDir)
	if os.IsNotExist(err) {
		fmt.Printf("crls folder not found at [%s]. Skipping. [%s]", crlsDir, err)
	} else if err != nil {
		return nil, fmt.Errorf("failed loading crls at [%s] %w", crlsDir, err)
	}

	// Load configuration file
	// if the configuration file is there then load it
	// otherwise skip it
	var ouis []*msp.FabricOUIdentifier
	var nodeOUs *msp.FabricNodeOUs
	_, err = os.Stat(configFile)
	if err == nil {
		// load the file, if there is a failure in loading it then
		// return an error
		raw, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed loading configuration file at [%s] %w", configFile, err)
		}

		configuration := Configuration{}
		err = yaml.Unmarshal(raw, &configuration)
		if err != nil {
			return nil, fmt.Errorf("failed unmarshalling configuration file at [%s] %w", configFile, err)
		}

		// Prepare OrganizationalUnitIdentifiers
		if len(configuration.OrganizationalUnitIdentifiers) > 0 {
			for _, ouID := range configuration.OrganizationalUnitIdentifiers {
				f := filepath.Join(dir, ouID.Certificate)
				raw, err = readFile(f)
				if err != nil {
					return nil, fmt.Errorf("failed loading OrganizationalUnit certificate at [%s] %w", f, err)
				}

				oui := &msp.FabricOUIdentifier{
					Certificate:                  raw,
					OrganizationalUnitIdentifier: ouID.OrganizationalUnitIdentifier,
				}
				ouis = append(ouis, oui)
			}
		}

		// Prepare NodeOUs
		if configuration.NodeOUs != nil && configuration.NodeOUs.Enable {
			fmt.Printf("Loading NodeOUs")
			nodeOUs = &msp.FabricNodeOUs{
				Enable: true,
			}
			if configuration.NodeOUs.ClientOUIdentifier != nil && len(configuration.NodeOUs.ClientOUIdentifier.OrganizationalUnitIdentifier) != 0 {
				nodeOUs.ClientOuIdentifier = &msp.FabricOUIdentifier{OrganizationalUnitIdentifier: configuration.NodeOUs.ClientOUIdentifier.OrganizationalUnitIdentifier}
			}
			if configuration.NodeOUs.PeerOUIdentifier != nil && len(configuration.NodeOUs.PeerOUIdentifier.OrganizationalUnitIdentifier) != 0 {
				nodeOUs.PeerOuIdentifier = &msp.FabricOUIdentifier{OrganizationalUnitIdentifier: configuration.NodeOUs.PeerOUIdentifier.OrganizationalUnitIdentifier}
			}
			if configuration.NodeOUs.AdminOUIdentifier != nil && len(configuration.NodeOUs.AdminOUIdentifier.OrganizationalUnitIdentifier) != 0 {
				nodeOUs.AdminOuIdentifier = &msp.FabricOUIdentifier{OrganizationalUnitIdentifier: configuration.NodeOUs.AdminOUIdentifier.OrganizationalUnitIdentifier}
			}
			if configuration.NodeOUs.OrdererOUIdentifier != nil && len(configuration.NodeOUs.OrdererOUIdentifier.OrganizationalUnitIdentifier) != 0 {
				nodeOUs.OrdererOuIdentifier = &msp.FabricOUIdentifier{OrganizationalUnitIdentifier: configuration.NodeOUs.OrdererOUIdentifier.OrganizationalUnitIdentifier}
			}

			// Read certificates, if defined

			// ClientOU
			if nodeOUs.GetClientOuIdentifier() != nil {
				nodeOUs.ClientOuIdentifier.Certificate = loadCertificateAt(dir, configuration.NodeOUs.ClientOUIdentifier.Certificate, "ClientOU")
			}
			// PeerOU
			if nodeOUs.GetPeerOuIdentifier() != nil {
				nodeOUs.PeerOuIdentifier.Certificate = loadCertificateAt(dir, configuration.NodeOUs.PeerOUIdentifier.Certificate, "PeerOU")
			}
			// AdminOU
			if nodeOUs.GetAdminOuIdentifier() != nil {
				nodeOUs.AdminOuIdentifier.Certificate = loadCertificateAt(dir, configuration.NodeOUs.AdminOUIdentifier.Certificate, "AdminOU")
			}
			// OrdererOU
			if nodeOUs.GetOrdererOuIdentifier() != nil {
				nodeOUs.OrdererOuIdentifier.Certificate = loadCertificateAt(dir, configuration.NodeOUs.OrdererOUIdentifier.Certificate, "OrdererOU")
			}
		}
	} else {
		fmt.Printf("MSP configuration file not found at [%s]: [%s]", configFile, err)
	}

	// Set FabricCryptoConfig
	cryptoConfig := &msp.FabricCryptoConfig{
		SignatureHashFamily:            "SHA2",
		IdentityIdentifierHashFunction: "SHA256",
	}

	// Compose FabricMSPConfig
	fmspconf := &msp.FabricMSPConfig{
		Admins:                        admincert,
		RootCerts:                     cacerts,
		IntermediateCerts:             intermediatecerts,
		SigningIdentity:               sigid,
		Name:                          ID,
		OrganizationalUnitIdentifiers: ouis,
		RevocationList:                crls,
		CryptoConfig:                  cryptoConfig,
		TlsRootCerts:                  tlsCACerts,
		TlsIntermediateCerts:          tlsIntermediateCerts,
		FabricNodeOus:                 nodeOUs,
	}

	fmpsjs, _ := proto.Marshal(fmspconf)

	mspconf := &msp.MSPConfig{Config: fmpsjs, Type: int32(FABRIC)}

	return mspconf, nil
}

func getPemMaterialFromDir(dir string) ([][]byte, error) {
	fmt.Printf("Reading directory %s", dir)

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, err
	}

	content := make([][]byte, 0)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory %s %w", dir, err)
	}

	for _, f := range files {
		fullName := filepath.Join(dir, f.Name())

		f, err := os.Stat(fullName)
		if err != nil {
			fmt.Printf("Failed to stat %s: %s", fullName, err)
			continue
		}
		if f.IsDir() {
			continue
		}

		fmt.Printf("Inspecting file %s", fullName)

		item, err := readPemFile(fullName)
		if err != nil {
			fmt.Printf("Failed reading file %s: %s", fullName, err)
			continue
		}

		content = append(content, item)
	}

	return content, nil
}

// OrganizationalUnitIdentifiersConfiguration is used to represent an OU
// and an associated trusted certificate
type OrganizationalUnitIdentifiersConfiguration struct {
	// Certificate is the path to a root or intermediate certificate
	Certificate string `yaml:"Certificate,omitempty"`
	// OrganizationalUnitIdentifier is the name of the OU
	OrganizationalUnitIdentifier string `yaml:"OrganizationalUnitIdentifier,omitempty"`
}

// NodeOUs contains information on how to tell apart clients, peers and orderers
// based on OUs. If the check is enforced, by setting Enabled to true,
// the MSP will consider an identity valid if it is an identity of a client, a peer or
// an orderer. An identity should have only one of these special OUs.
type NodeOUs struct {
	// Enable activates the OU enforcement
	Enable bool `yaml:"Enable,omitempty"`
	// ClientOUIdentifier specifies how to recognize clients by OU
	ClientOUIdentifier *OrganizationalUnitIdentifiersConfiguration `yaml:"ClientOUIdentifier,omitempty"`
	// PeerOUIdentifier specifies how to recognize peers by OU
	PeerOUIdentifier *OrganizationalUnitIdentifiersConfiguration `yaml:"PeerOUIdentifier,omitempty"`
	// AdminOUIdentifier specifies how to recognize admins by OU
	AdminOUIdentifier *OrganizationalUnitIdentifiersConfiguration `yaml:"AdminOUIdentifier,omitempty"`
	// OrdererOUIdentifier specifies how to recognize admins by OU
	OrdererOUIdentifier *OrganizationalUnitIdentifiersConfiguration `yaml:"OrdererOUIdentifier,omitempty"`
}

// Configuration represents the accessory configuration an MSP can be equipped with.
// By default, this configuration is stored in a yaml file
type Configuration struct {
	// OrganizationalUnitIdentifiers is a list of OUs. If this is set, the MSP
	// will consider an identity valid only it contains at least one of these OUs
	OrganizationalUnitIdentifiers []*OrganizationalUnitIdentifiersConfiguration `yaml:"OrganizationalUnitIdentifiers,omitempty"`
	// NodeOUs enables the MSP to tell apart clients, peers and orderers based
	// on the identity's OU.
	NodeOUs *NodeOUs `yaml:"NodeOUs,omitempty"`
}

func readFile(file string) ([]byte, error) {
	fileCont, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s %w", file, err)
	}

	return fileCont, nil
}

func readPemFile(file string) ([]byte, error) {
	bytes, err := readFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading from file %s failed %w", file, err)
	}

	b, _ := pem.Decode(bytes)
	if b == nil { // TODO: also check that the type is what we expect (cert vs key..)
		return nil, fmt.Errorf("no pem content for file %s", file)
	}

	return bytes, nil
}

func loadCertificateAt(dir, certificatePath string, ouType string) []byte {
	f := filepath.Join(dir, certificatePath)
	raw, err := readFile(f)
	if err != nil {
		fmt.Printf("Failed loading %s certificate at [%s]: [%s]", ouType, f, err)
	} else {
		return raw
	}

	return nil
}

const (
	IdemixConfigDirMsp                  = "msp"
	IdemixConfigDirUser                 = "user"
	IdemixConfigFileIssuerPublicKey     = "IssuerPublicKey"
	IdemixConfigFileRevocationPublicKey = "RevocationPublicKey"
	IdemixConfigFileSigner              = "SignerConfig"
)

// GetIdemixMspConfig returns the configuration for the Idemix MSP
func GetIdemixMspConfig(dir string, ID string) (*msp.MSPConfig, error) {
	ipkBytes, err := readFile(filepath.Join(dir, IdemixConfigDirMsp, IdemixConfigFileIssuerPublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to read issuer public key file %w", err)
	}

	revocationPkBytes, err := readFile(filepath.Join(dir, IdemixConfigDirMsp, IdemixConfigFileRevocationPublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to read revocation public key file %w", err)
	}

	idemixConfig := &msp.IdemixMSPConfig{
		Name:         ID,
		Ipk:          ipkBytes,
		RevocationPk: revocationPkBytes,
	}

	signerBytes, err := readFile(filepath.Join(dir, IdemixConfigDirUser, IdemixConfigFileSigner))
	if err == nil {
		signerConfig := &msp.IdemixMSPSignerConfig{}
		err = proto.Unmarshal(signerBytes, signerConfig)
		if err != nil {
			return nil, err
		}
		idemixConfig.Signer = signerConfig
	}

	confBytes, err := proto.Marshal(idemixConfig)
	if err != nil {
		return nil, err
	}

	return &msp.MSPConfig{Config: confBytes, Type: int32(IDEMIX)}, nil
}

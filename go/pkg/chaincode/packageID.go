package chaincode

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/hyperledger/fabric/common/util"
	"github.com/pkg/errors"
)

const (
	// MetadataFile is the expected location of the metadata json document
	// in the top level of the chaincode package.
	MetadataFile = "metadata.json"

	// CodePackageFile is the expected location of the code package in the
	// top level of the chaincode package
	CodePackageFile = "code.tar.gz"
)

var LabelRegexp = regexp.MustCompile(`^[[:alnum:]][[:alnum:]_.+-]*$`)

func PackageID(PackageFile string) (string, error) {
	pkgBytes, err := os.ReadFile(PackageFile)
	if err != nil {
		return "", errors.WithMessagef(err, "failed to read chaincode package at '%s'", PackageFile)
	}
	metadata, _, err := ParseChaincodePackage(pkgBytes)
	if err != nil {
		return "", errors.WithMessage(err, "could not parse as a chaincode install package")
	}
	packageID := GetPackageID(metadata.Label, pkgBytes)
	return packageID, nil
}

// PackageID returns the package ID with the label and hash of the chaincode install package
func GetPackageID(label string, ccInstallPkg []byte) string {
	hash := util.ComputeSHA256(ccInstallPkg)
	return fmt.Sprintf("%s:%x", label, hash)
}

// ChaincodePackageMetadata contains the information necessary to understand
// the embedded code package.
type ChaincodePackageMetadata struct {
	Type  string `json:"type"`
	Path  string `json:"path"`
	Label string `json:"label"`
}

// ParseChaincodePackage parses a set of bytes as a chaincode package
// and returns the parsed package as a metadata struct and a code package
func ParseChaincodePackage(source []byte) (*ChaincodePackageMetadata, []byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewBuffer(source))
	if err != nil {
		return &ChaincodePackageMetadata{}, nil, errors.Wrapf(err, "error reading as gzip stream")
	}

	tarReader := tar.NewReader(gzReader)

	var codePackage []byte
	var ccPackageMetadata *ChaincodePackageMetadata
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return ccPackageMetadata, nil, errors.Wrapf(err, "error inspecting next tar header")
		}

		if header.Typeflag != tar.TypeReg {
			return ccPackageMetadata, nil, errors.Errorf("tar entry %s is not a regular file, type %v", header.Name, header.Typeflag)
		}

		fileBytes, err := io.ReadAll(tarReader)
		if err != nil {
			return ccPackageMetadata, nil, errors.Wrapf(err, "could not read %s from tar", header.Name)
		}

		switch header.Name {

		case MetadataFile:
			ccPackageMetadata = &ChaincodePackageMetadata{}
			err := json.Unmarshal(fileBytes, ccPackageMetadata)
			if err != nil {
				return ccPackageMetadata, nil, errors.Wrapf(err, "could not unmarshal %s as json", MetadataFile)
			}

		case CodePackageFile:
			codePackage = fileBytes
		default:
			fmt.Println("Encountered unexpected file " + header.Name + " in top level of chaincode package")
		}
	}

	if codePackage == nil {
		return ccPackageMetadata, nil, errors.Errorf("did not find a code package inside the package")
	}

	if ccPackageMetadata == nil {
		return ccPackageMetadata, nil, errors.Errorf("did not find any package metadata (missing %s)", MetadataFile)
	}

	if err := ValidateLabel(ccPackageMetadata.Label); err != nil {
		return ccPackageMetadata, nil, err
	}

	return ccPackageMetadata, codePackage, nil
}

// ValidateLabel return an error if the provided label contains any invalid
// characters, as determined by LabelRegexp.
func ValidateLabel(label string) error {
	if !LabelRegexp.MatchString(label) {
		return errors.Errorf("invalid label '%s'. Label must be non-empty, can only consist of alphanumerics, symbols from '.+-_', and can only begin with alphanumerics", label)
	}

	return nil
}

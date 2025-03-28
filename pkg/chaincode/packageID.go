package chaincode

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

func PackageID(packageReader io.Reader) (string, error) {
	pkgBytes, err := io.ReadAll(packageReader)
	if err != nil {
		return "", fmt.Errorf("failed to read chaincode package: %w", err)
	}
	metadata, _, err := ParseChaincodePackage(pkgBytes)
	if err != nil {
		return "", fmt.Errorf("could not parse as a chaincode install package %w", err)
	}
	packageID := GetPackageID(metadata.Label, pkgBytes)
	return packageID, nil
}

// GetPackageID returns the package ID with the label and hash of the chaincode install package
func GetPackageID(label string, ccInstallPkg []byte) string {
	h := sha256.New()
	h.Write(ccInstallPkg)
	hash := h.Sum(nil)
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
//
//nolint:cyclop,gocognit
func ParseChaincodePackage(source []byte) (*ChaincodePackageMetadata, []byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewBuffer(source))
	if err != nil {
		return &ChaincodePackageMetadata{}, nil, fmt.Errorf("error reading as gzip stream %w", err)
	}

	tarReader := tar.NewReader(gzReader)

	var codePackage []byte
	var ccPackageMetadata *ChaincodePackageMetadata
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return ccPackageMetadata, nil, fmt.Errorf("error inspecting next tar header %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			return ccPackageMetadata, nil, fmt.Errorf("tar entry %s is not a regular file, type %v %w", header.Name, header.Typeflag, err)
		}

		fileBytes, err := io.ReadAll(tarReader)
		if err != nil {
			return ccPackageMetadata, nil, fmt.Errorf("could not read %s from tar %w", header.Name, err)
		}

		switch header.Name {

		case metadataFile:
			ccPackageMetadata = &ChaincodePackageMetadata{}
			err := json.Unmarshal(fileBytes, ccPackageMetadata)
			if err != nil {
				return ccPackageMetadata, nil, fmt.Errorf("could not unmarshal %s as json %w", metadataFile, err)
			}

		case codePackageFile:
			codePackage = fileBytes
		default:
			fmt.Println("Encountered unexpected file " + header.Name + " in top level of chaincode package")
		}
	}

	if codePackage == nil {
		return ccPackageMetadata, nil, fmt.Errorf("did not find a code package inside the package")
	}

	if ccPackageMetadata == nil {
		return ccPackageMetadata, nil, fmt.Errorf("did not find any package metadata (missing %s)", metadataFile)
	}

	if err := ValidateLabel(ccPackageMetadata.Label); err != nil {
		return ccPackageMetadata, nil, err
	}

	return ccPackageMetadata, codePackage, nil
}

// ValidateLabel return an error if the provided label contains any invalid
// characters, as determined by LabelRegexp.
func ValidateLabel(label string) error {
	LabelRegexp := regexp.MustCompile(`^[[:alnum:]][[:alnum:]_.+-]*$`)
	if !LabelRegexp.MatchString(label) {
		return fmt.Errorf("invalid label '%s'. Label must be non-empty, can only consist of alphanumerics, symbols from '.+-_', and can only begin with alphanumerics", label)
	}

	return nil
}

package chaincode

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Package", func() {
	It("CCaaS", func() {
		dummyConnection := Connection{
			Address:      "127.0.0.1:8080",
			Dial_timeout: "10s",
			Tls_required: false,
		}
		dummyMeta := Metadata{
			Type:  "ccaas",
			Label: "basic-asset",
		}
		err := PackageCCAAS(dummyConnection, dummyMeta, tmpDir, "chaincode.tar.gz")
		Expect(err).NotTo(HaveOccurred())
		// so far no plan to verify the file
		file, err := os.Open(tmpDir + "/chaincode.tar.gz")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()
		err = Untar(file, tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})
})

// Untar takes a gzip-ed tar archive, and extracts it to dst.
// It returns an error if the tar contains any files which would escape to a
// parent of dst, or if the archive contains any files whose type is not
// a regular file or directory.
func Untar(buffer io.Reader, dst string) error {
	gzr, err := gzip.NewReader(buffer)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("could not get next tar element %w", err)
		}

		target, err := SanitizeArchivePath(dst, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o700); err != nil {
				return fmt.Errorf("could not create directory '%s' %w", header.Name, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return fmt.Errorf("could not create directory '%s' %w", filepath.Dir(header.Name), err)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("could not create file '%s' %w", header.Name, err)
			}

			// copy over contents
			//#nosec G110 - Leave to the user to check for decompression bomb?
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		default:
			return fmt.Errorf("invalid file type '%v' contained in archive for file '%s'", header.Typeflag, header.Name)
		}
	}
}

func SanitizeArchivePath(d, t string) (string, error) {
	target := filepath.Join(d, t)
	if !strings.HasPrefix(target, filepath.Clean(d)) {
		return "", fmt.Errorf("content filepath is tainted: %s", t)
	}

	return target, nil
}

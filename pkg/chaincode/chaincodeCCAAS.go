package chaincode

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Connection setting used in connection.json
type Connection struct {
	Address     string `json:"address"`
	DialTimeout string `json:"dial_timeout"`
	TLSRequired bool   `json:"tls_required"`
}

// Metadata as metadata.json
type Metadata struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

/*
PackageCCAAS requires that you start a container by yourself. as sample as
${CONTAINER_CLI} run --rm -d --name peer0org1_${CC_NAME}_ccaas  \
--network fabric_test \
-e CHAINCODE_SERVER_ADDRESS=0.0.0.0:${CCAAS_SERVER_PORT} \
-e CHAINCODE_ID=$PACKAGE_ID -e CORE_CHAINCODE_ID_NAME=$PACKAGE_ID \
${CC_NAME}_ccaas_image:latest
*/
func PackageCCAAS(connection Connection, metadata Metadata, tmpPath, filename string) error {
	if err := os.MkdirAll(tmpPath+"/src/", 0766); err != nil {
		return err
	}

	if err := os.MkdirAll(tmpPath+"/pkg/", 0766); err != nil {
		return err
	}

	if err := writeJSONFile(connection, tmpPath+"/src/connection.json"); err != nil {
		return err
	}

	// code.tar.gz
	if err := creategzTar(tmpPath+"/pkg/"+"code.tar.gz", []string{tmpPath + "/src/connection.json"}); err != nil {
		return err
	}

	// metadata.json
	if err := writeJSONFile(metadata, tmpPath+"/pkg/metadata.json"); err != nil {
		return err
	}

	//filename
	if err := creategzTar(tmpPath+"/"+filename, []string{tmpPath + "/pkg/metadata.json", tmpPath + "/pkg/code.tar.gz"}); err != nil {
		return err
	}

	return nil
}

func writeJSONFile(jsonObject any, filePath string) error {
	json, err := json.Marshal(jsonObject)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(json)
	if err != nil {
		return err
	}

	return nil
}

func creategzTar(filename string, fileList []string) error {
	fw, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fw.Close()
	gw := gzip.NewWriter(fw)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, v := range fileList {
		f, err := os.Open(v)
		if err != nil {
			fmt.Printf("open %s err:%s\n", v, err.Error())
			continue
		}

		info, _ := f.Stat()
		header, _ := tar.FileInfoHeader(info, "")

		err = tw.WriteHeader(header)
		if err != nil {
			f.Close()
			return err
		}
		_, err = io.Copy(tw, f)
		if err != nil {
			return err
		}
		f.Close()
	}

	return nil
}

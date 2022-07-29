package chaincode

import (
	"archive/tar"
	"bytes"
	"encoding/json"
)

type Connection struct {
	Address      string `json:"address"`
	Dial_timeout string `json:"dial_timeout"`
	Tls_required bool   `json:"tls_required"`
}

type Metadata struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

type TarFile struct {
	Name string
	Body string
}

/*
To use PackageCCAAS you need to start container by yourself. as sample as
${CONTAINER_CLI} run --rm -d --name peer0org1_${CC_NAME}_ccaas  \
--network fabric_test \
-e CHAINCODE_SERVER_ADDRESS=0.0.0.0:${CCAAS_SERVER_PORT} \
-e CHAINCODE_ID=$PACKAGE_ID -e CORE_CHAINCODE_ID_NAME=$PACKAGE_ID \
${CC_NAME}_ccaas_image:latest
*/
func PackageCCAAS(connection Connection, metadata Metadata) (string, error) {
	connjsonStr, err := json.Marshal(connection)
	if err != nil {
		return "", err
	}
	connectionTarFile := TarFile{
		Name: "connection.json",
		Body: string(connjsonStr),
	}
	codeTarString, err := createTar([]TarFile{connectionTarFile})
	if err != nil {
		return "", err
	}
	metajsonStr, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	metaTarFile := TarFile{
		Name: "metadata.json",
		Body: string(metajsonStr),
	}
	codeTarFile := TarFile{
		Name: "code.tar.gz",
		Body: codeTarString,
	}
	//tar -C "$tempdir/pkg" -czf "$CC_NAME.tar.gz" metadata.json code.tar.gz
	finalString, err := createTar([]TarFile{codeTarFile, metaTarFile})
	if err != nil {
		return "", err
	}
	return finalString, nil
}

func createTar(files []TarFile) (string, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return "", err
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

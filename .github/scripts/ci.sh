#!/usr/bin/env bash
#

set -eu -o pipefail

DIR="${PWD}"

curl --silent --show-error --location https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh | bash -s -- --fabric-version "$1"

cd fabric-samples/test-network
if [[ $2 == 'create_channel' ]]; then
    ./network.sh up
else    
    ./network.sh up createChannel -c mychannel
fi

cd "${DIR}/test"
mkdir -p ../organizations
cp -r ../fabric-samples/test-network/organizations/* ../organizations/
ls ../organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
if [[ $2 == 'create_channel' ]]; then
    export createChannel=true
fi
go test -v ./...
docker network ls
docker ps -a
cat PackageID
docker run --rm -d --name peer0org1_basic --network fabric_test -e CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999 -e CORE_CHAINCODE_ID_NAME="$(cat PackageID)" ghcr.io/hyperledgendary/fabric-ccaas-asset-transfer-basic:latest
docker run --rm -d --name peer0org2_basic --network fabric_test -e CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999 -e CORE_CHAINCODE_ID_NAME="$(cat PackageID)" ghcr.io/hyperledgendary/fabric-ccaas-asset-transfer-basic:latest

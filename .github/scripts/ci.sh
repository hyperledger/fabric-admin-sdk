#!/usr/bin/env bash
#

set -eu -o pipefail

DIR="${PWD}"

curl --silent --show-error --location https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh | bash -s -- --fabric-version "$FABRIC_VERSION"
cd fabric-samples/test-network

echo $FABRIC_VERSION
echo $CONSENSUS

if [[ $CONSENSUS == 'BFT' ]]; then 
    echo start 3.0.0 preview BFT processing
    if [[ $CREATE_CHANNEL == 'create_channel' ]]; then
        ./network.sh up -bft
    else    
        echo y | ./network.sh up createChannel -bft -c mychannel
    fi
else
    if [[ $CREATE_CHANNEL == 'create_channel' ]]; then
        ./network.sh up
    else    
        ./network.sh up createChannel -c mychannel
    fi
fi

cd "${DIR}/test"
mkdir -p ../organizations
cp -r ../fabric-samples/test-network/organizations/* ../organizations/
ls ../organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt

go test -v ./...
docker network ls
docker ps -a
cat PackageID
docker run --rm -d --name peer0org1_basic --network fabric_test -e CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999 -e CORE_CHAINCODE_ID_NAME="$(cat PackageID)" ghcr.io/hyperledgendary/fabric-ccaas-asset-transfer-basic:latest
docker run --rm -d --name peer0org2_basic --network fabric_test -e CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999 -e CORE_CHAINCODE_ID_NAME="$(cat PackageID)" ghcr.io/hyperledgendary/fabric-ccaas-asset-transfer-basic:latest

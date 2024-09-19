# Fabric Admin SDK

The fabric-admin-sdk provides administrative APIs for creating and configuring [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/) networks, and for deploying smart contracts (chaincode) to those networks.

For information on using the fabric-admin-sdk, please visit the [Go API documentation](https://pkg.go.dev/github.com/hyperledger/fabric-admin-sdk/pkg). The API documentation includes some usage examples. Additionally, the [scenario tests](./test/e2e_test.go) can be used as an example of an end-to-end flow, covering channel creation and smart contract deployment.

## Overview

The current API for client applications to transact with Fabric networks (the [Fabric Gateway client API](https://hyperledger.github.io/fabric-gateway/)) does not provide the administrative capabilities of the legacy client application SDKs. The fabric-admin-sdk aims to provide this programmatic administrative capability for use-cases where the [Fabric CLI commands](https://hyperledger-fabric.readthedocs.io/en/latest/command_ref.html) are not the best fit. A specific objective is to support the development of Kubernetes operators.

More detailed information on the motivation and objectives of the fabric-admin-sdk can be found in the associated [Fabric Admin SDK RFC](https://hyperledger.github.io/fabric-rfcs/text/0000-admin_sdk.html).

## Building and testing

### Install pre-reqs

This repository contains APIs written in Go and JavaScript. In order to build these components, the following need to be installed and available in the PATH:

- [Go 1.21+](https://go.dev/)
- [Node 18+](https://nodejs.org/)
- [Make](https://www.gnu.org/software/make/)

In order to run scenario test, [Docker](https://www.docker.com/) is also required.

#### Dev Container

This project includes a [Dev Container](https://containers.dev/) configuration that includes all of the pre-requisite software described above in a Docker container, avoiding the need to install them locally. The only requirement is that [Docker](https://www.docker.com/) is installed and available.

Opening the project folder in an IDE such as [VS Code](https://code.visualstudio.com/docs/devcontainers/containers) (with the [Dev Containers extention](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)) should offer the option of opening in the Dev Container. Alternatively, VS Code allows the remote repository to [opened directly in an isolated Dev Container](https://code.visualstudio.com/docs/devcontainers/containers#_quick-start-open-a-git-repository-or-github-pr-in-an-isolated-container-volume).

### Build using make

> **Note:** When the repository is first cloned, some mock implementations used for testing will not be present and the Go code will show compile errors. These can be generated explicitly by running `make generate`.

The following Makefile targets are available:

- `make generate` - generate mock implementations used by unit tests
- `make lint` - run linting checks for the Go code
- `make unit-test-go` - run unit tests for the Go API
- `make unit-test-node` - run unit tests for the Node API
- `make unit-test` - run unit tests for all language implementations

## History

The initial submission and implementation of fabric-admin-sdk was driven by members of the [Technical Working Group China](https://github.com/Hyperledger-TWGC) / 超级账本中国技术工作组 (TWGC):

- [davidkhala](https://github.com/davidkhala)
- [SamYuan1990](https://github.com/SamYuan1990)
- [xiaohui249](https://github.com/xiaohui249)
- [Peng-Du](https://github.com/Peng-Du)

## Contribute

Here are steps in short for any contribution.

1. check license and code of conduct
1. fork this project
1. make your own feature branch
1. change and commit your changes, please use `git commit -s` to commit as we enabled [DCO](https://probot.github.io/apps/dco/)
1. raise PR

## Code of Conduct guidelines

Please review the Hyperledger [Code of Conduct](https://wiki.hyperledger.org/community/hyperledger-project-code-of-conduct) before participating. It is important that we keep things civil.

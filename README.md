# fabric-admin-sdk
Fabric SDK for Admin Capability services 

## [motivation](https://github.com/hyperledger/fabric-rfcs/pull/55)
As gateway sdk will drop off admin capacity, we plan to recover admin capacity related things with this project proposal.

## identified [features](https://github.com/Hyperledger-TWGC/fabric-admin-sdk/issues/15)

### channel
- new channel
- channel update(new org join into specific channel)
- peer join channel
- peer exit channel
- list channel 
- inspect channel config

### chaincode
- chain code lifecycle
- system chain code(as list all chain code)

### tools
- gate policy?(common/policydsl/policyparser.go, tool for operator)
- peer discovery(optional, as ping test among peer networks when gateway disabled or fabric version below 2.5)
- base on peer discovery result, generate connection profile for sdk if possible?(optional)

## languages: go, java, nodejs, typescript
### [golang Documentation](https://pkg.go.dev/github.com/hyperledger/fabric-admin-sdk)

## init by 
- [davidkhala](https://github.com/davidkhala)
- [SamYuan1990](https://github.com/SamYuan1990)
- [xiaohui249](https://github.com/xiaohui249)
- [Peng-Du](https://github.com/Peng-Du)

## Contributors
<a href="https://github.com/Hyperledger/fabric-admin-sdk/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=Hyperledger/fabric-admin-sdk" />
</a>

### [Scenario tests and example](./test/e2e_test.go)

## Contribute
Here is steps in short for any contribution. 
1. check license and code of conduct
1. fork this project
1. make your own feature branch
1. change and commit your changes, please use `git commit -s` to commit as we enabled [DCO](https://probot.github.io/apps/dco/)
1. raise PR

## Code of Conduct guidelines
Please review the Hyperledger [Code of
Conduct](https://wiki.hyperledger.org/community/hyperledger-project-code-of-conduct)
before participating. It is important that we keep things civil.

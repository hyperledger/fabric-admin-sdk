import {getNonce, calculateTransactionId} from '../lib/helper'
import {buildHeader, buildChannelHeader} from '../lib/proto/channel-builder'
import {buildSerializedIdentity, buildSignatureHeader,} from "../lib/proto/common-builder";
import {ChannelHeader, HeaderType} from "@hyperledger/fabric-protos/lib/common/common_pb";
import {decode} from '../lib/proto/common-parser'

const certificate = `
        -----BEGIN CERTIFICATE-----
MIICmDCCAj+gAwIBAgIUIJvfcIDuIjyq/ugxhJuplP6QBN0wCgYIKoZIzj0EAwIw
aDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMRQwEgYDVQQK
EwtIeXBlcmxlZGdlcjEPMA0GA1UECxMGRmFicmljMRkwFwYDVQQDExBmYWJyaWMt
Y2Etc2VydmVyMB4XDTIyMDYxNzA4NTcwMFoXDTIzMDYxNzA5MDIwMFowQzErMAsG
A1UECxMEaWNkZDANBgNVBAsTBmNsaWVudDANBgNVBAsTBmNsaWVudDEUMBIGA1UE
AxMLaWNkZC5jbGllbnQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATh3c8TXEkP
JuXjVQvwic0eGoE3TPmglgOUq0nLygqykAGV/QktQ1Lfp0X4cdhIXCkXSUMphBh+
KcedLPRIpsCxo4HrMIHoMA4GA1UdDwEB/wQEAwIDqDAdBgNVHSUEFjAUBggrBgEF
BQcDAQYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUTuzZdW2/tBa/
F1vmE99b7evs/mowHwYDVR0jBBgwFoAUM4k1teT9ElsxoYgEaSi260EyXAMwaQYI
KgMEBQYHCAEEXXsiYXR0cnMiOnsiaGYuQWZmaWxpYXRpb24iOiJpY2RkLmNsaWVu
dCIsImhmLkVucm9sbG1lbnRJRCI6ImljZGQuY2xpZW50IiwiaGYuVHlwZSI6ImNs
aWVudCJ9fTAKBggqhkjOPQQDAgNHADBEAiBZ8VplC5jr1Y7vm7+Zc4bz6gcrzIlw
n8i/3IY/tFLRdAIgYK0FrVOE5dup/acc5oaRkagZ4bBN84vtwym4Y924D2I=
-----END CERTIFICATE-----
`
describe('type test', () => {
    const mspid = 'org1.msp'
    const nonce = getNonce()
    it('buildSignatureHeader', () => {
        const creator = buildSerializedIdentity({mspid, idBytes: Buffer.from(certificate)}).serializeBinary()
        const signatureHeader = buildSignatureHeader({creator, nonce})

    })
    it('buildChannelHeader', () => {

        const type = HeaderType.CONFIG
        const channel = 'mychannel'
        const creator = buildSerializedIdentity({mspid,idBytes:Buffer.from(certificate)}).serializeBinary()
        const txid = calculateTransactionId({creator, nonce})
        const channelHeader = buildChannelHeader(type, channel, txid)
        const bytes = channelHeader.serializeBinary()
        const obj = decode(bytes, ChannelHeader)
        console.debug(obj)
    })
})
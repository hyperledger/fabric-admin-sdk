import {SerializedIdentity} from "@hyperledger/fabric-protos/lib/msp/identities_pb";

export interface IdentityContext extends SerializedIdentity.AsObject{
    sign(payload: Uint8Array): Promise<Uint8Array>;
}
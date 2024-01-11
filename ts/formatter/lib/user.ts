import {msp} from "@hyperledger/fabric-protos";

export interface IdentityContext extends msp.SerializedIdentity.AsObject{
    sign(payload: Uint8Array): Promise<Uint8Array>;
}
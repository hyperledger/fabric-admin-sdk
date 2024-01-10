import {msp} from '@hyperledger/fabric-protos';
import {MspId} from "./types";
import MSPRoleTypeMap = msp.MSPRole.MSPRoleTypeMap;

const {MSPRole, MSPPrincipal} = msp


export const Classification = [
    'ROLE',  // Represents the one of the dedicated MSP roles, the one of a member of MSP network, and the one of an administrator of an MSP network
    'ORGANIZATION_UNIT', // Denotes a finer grained (affiliation-based) grouping of entities, per MSP affiliation     // E.g., this can well be represented by an MSP's Organization unit
    'IDENTITY',   // Denotes a principal that consists of a single identity
    'ANONYMITY', // Denotes a principal that can be used to enforce an identity to be anonymous or nominal.
    'COMBINED' // Denotes a combined principal
]


export function build(role: MSPRoleTypeMap[keyof MSPRoleTypeMap], mspId: MspId, classification = MSPPrincipal.Classification.ROLE) {
    const newPrincipal = new MSPPrincipal();
    newPrincipal.setPrincipalClassification(classification);
    const newRole = new MSPRole()
    newRole.setRole(role)
    newRole.setMspIdentifier(mspId)
    newPrincipal.setPrincipal(newRole.serializeBinary());
    return newPrincipal;
}
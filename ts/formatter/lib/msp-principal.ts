import {msp} from '@hyperledger/fabric-protos';
import {MspId} from "./d";

const {MSPPrincipal, MSPRole} = msp

export const Classification = [
    'ROLE',  // Represents the one of the dedicated MSP roles, the one of a member of MSP network, and the one of an administrator of an MSP network
    'ORGANIZATION_UNIT', // Denotes a finer grained (affiliation-based) grouping of entities, per MSP affiliation     // E.g., this can well be represented by an MSP's Organization unit
    'IDENTITY',   // Denotes a principal that consists of a single identity
    'ANONYMITY', // Denotes a principal that can be used to enforce an identity to be anonymous or nominal.
    'COMBINED' // Denotes a combined principal
]

/**
 *
 * @param {MSPRoleTypeMap[keyof MSPRoleTypeMap]} MSPRoleType
 * @param {MspId} mspid
 * @param {ClassificationMap} [classification]
 * @return {MSPPrincipal}
 */
export function build(MSPRoleType, mspid:MspId, classification = MSPPrincipal.Classification.ROLE) {
    const newPrincipal = new MSPPrincipal();
    newPrincipal.setPrincipalClassification(classification);
    const newRole = new MSPRole()
    newRole.setRole(MSPRoleType)
    newRole.setMspIdentifier(mspid)
    newPrincipal.setPrincipal(newRole.serializeBinary());
    return newPrincipal;
}
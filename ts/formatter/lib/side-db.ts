import {common, msp} from "@hyperledger/fabric-protos";
import * as MSPPrincipalNS from './msp-principal'
import * as SignaturePolicyNS from './signature-policy'

const {MSPRole} = msp
const {MSPRoleType} = MSPRole;
const {
    CollectionConfig,
    CollectionPolicyConfig,
    SignaturePolicyEnvelope,
    StaticCollectionConfig,
    ApplicationPolicy
} = common

export function implicitCollection(mspid) {
    return `_implicit_org_${mspid}`
}

export function buildCollectionConfig({
                                          name,
                                          requiredPeerCount,
                                          maxPeerCount = requiredPeerCount,
                                          endorsement_policy,
                                          blockToLive,
                                          memberOnlyRead = true,
                                          memberOnlyWrite = true,
                                          member_orgs
                                      }) {


    const collectionConfig = new CollectionConfig();

    // a reference to a policy residing / managed in the config block to define which orgs have access to this collection’s private data
    const collectionPolicyConfig = new CollectionPolicyConfig();
    const signaturePolicyEnvelope = new SignaturePolicyEnvelope();

    const identities = member_orgs.map(mspid => {
        return MSPPrincipalNS.build(MSPRoleType.MEMBER, mspid);
    });


    const rules = member_orgs.map((mspid, index) => {
        return SignaturePolicyNS.build({signed_by: index});
    });
    const n_out_of = SignaturePolicyNS.buildNOutOf({n: 1, rules});
    const rule = SignaturePolicyNS.build({n_out_of});
    signaturePolicyEnvelope.setRule(rule);
    signaturePolicyEnvelope.setIdentitiesList(identities);

    collectionPolicyConfig.setSignaturePolicy(signaturePolicyEnvelope);

    const staticCollectionConfig = new StaticCollectionConfig();
    staticCollectionConfig.setName(name)
    staticCollectionConfig.setRequiredPeerCount(requiredPeerCount)
    staticCollectionConfig.setMaximumPeerCount(maxPeerCount)
    if (blockToLive) {
        staticCollectionConfig.setBlockToLive(blockToLive);
    }

    staticCollectionConfig.setMemberOnlyWrite(memberOnlyWrite);
    staticCollectionConfig.setMemberOnlyRead(memberOnlyRead);

    staticCollectionConfig.setMemberOrgsPolicy(collectionPolicyConfig);
    if (endorsement_policy) {
        const {channel_config_policy_reference, signature_policy} = endorsement_policy;
        const applicationPolicy = new ApplicationPolicy();

        if (channel_config_policy_reference) {
            applicationPolicy.setChannelConfigPolicyReference(channel_config_policy_reference);
        } else if (signature_policy) {
            applicationPolicy.setSignaturePolicy(signature_policy);
        }
        staticCollectionConfig.setEndorsementPolicy(applicationPolicy);
    }

    collectionConfig.setStaticCollectionConfig(staticCollectionConfig);
    return collectionConfig;
}


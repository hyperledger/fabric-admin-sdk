/**
 *
 * @param {SigningIdentity} signingIdentity
 * @return {module:api.Key} publicKey The public key represented by the certificate
 */
export const getPublicKey = (signingIdentity) => signingIdentity._publicKey;


/**
 *
 * @param {SigningIdentity} signingIdentity
 * @return {ECDSA_KEY}
 */
export const getPrivateKey = (signingIdentity) => signingIdentity._signer._key;

/**
 *
 * @param {SigningIdentity} signingIdentity
 * @return {CertificatePem}
 */
export const getCertificate = (signingIdentity) => signingIdentity._certificate.toString().trim();

/**
 *
 * @param {SigningIdentity} signingIdentity
 * @return {MspId}
 */
export const getMSPID = (signingIdentity) => signingIdentity._mspId;

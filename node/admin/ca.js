import FabricCAClient from 'fabric-ca-client/lib/FabricCAClient.js';
import {ECDSAConfig, ECDSAKey} from '@davidkhala/crypto/ECDSA.js';
import {Extension} from '@davidkhala/crypto/extension.js';
import {emptySuite} from 'khala-fabric-admin/cryptoSuite.js';
import {asn1} from 'jsrsasign';

export default class FabricCAService {

	constructor({trustedRoots = [], protocol, hostname, port, caname = ''}, cryptoSuite = emptySuite(), logger = console) {
		const tlsOptions = {
			trustedRoots,
			verify: trustedRoots.length > 0
		};
		this._fabricCAClient = new FabricCAClient({
			caname,
			protocol,
			hostname,
			port,
			tlsOptions,
		}, cryptoSuite);
		Object.assign(this, {caname, _cryptoSuite: cryptoSuite, logger});
	}

	/**
	 * @typedef {Object} AttributeRequest
	 * @property {string} name - The name of the attribute to include in the certificate
	 * @property {boolean} optional - throw an error if the identity does not have the attribute
	 */

	/**
	 * @typedef {Object} EnrollmentRequest
	 * @property {string} enrollmentID - The registered ID to use for enrollment
	 * @property {string} enrollmentSecret - The secret associated with the enrollment ID
	 * @property {string} [profile] - The profile name.  Specify the 'tls' profile for a TLS certificate;
	 *                   otherwise, an enrollment certificate is issued.
	 * @property {string} [csr] - Optional. PEM-encoded PKCS#10 Certificate Signing Request.
	 * @property {AttributeRequest[]} [attr_reqs]
	 */

	/**
	 * Enroll the member and return an opaque member object.
	 *
	 * @param {EnrollmentRequest} req If the request contains the field "csr", this csr will be used for
	 *     getting the certificate from Fabric-CA. Otherwise , a new private key will be generated and be used to
	 *     generate a csr later.
	 * @param subject
	 * @param dns
	 */
	async enroll(req, {subject, dns = []} = {}) {
		const {enrollmentID, enrollmentSecret, profile, attr_reqs} = req;

		if (!subject) {
			subject = `CN=${enrollmentID}`;
		}
		let {csr} = req;

		const {_cryptoSuite} = this;
		const result = {};
		if (!csr) {
			const keySize = _cryptoSuite._keySize;
			const config = new ECDSAConfig(keySize);
			const keyPair = config.generateEphemeralKey();
			const key = new ECDSAKey(keyPair, {keySize});
			Object.assign(result, {keyPair, key: keyPair.prvKeyObj});
			const extension = Extension.asSAN(dns);
			csr = key.generateCSR({str: asn1.x509.X500Name.ldapToOneline(subject)}, [extension]);
		}

		const enrollResponse = await this._fabricCAClient.enroll(enrollmentID, enrollmentSecret, csr, profile, attr_reqs);

		Object.assign(result, {
			certificate: enrollResponse.enrollmentCert,
			rootCertificate: enrollResponse.caCertChain
		});
		return result;
	}

	// The Idemix credential issuance is a two step process.
	// First, send a request with an empty body to the /api/v1/idemix/credential API endpoint to get a nonce and CA’s Idemix public key.
	// Second, create a credential request using the nonce and CA’s Idemix public key and send another request with the credential request in the body to the /api/v1/idemix/credential API endpoint to get an Idemix credential,
	//      Credential Revocation Information (CRI), and attribute names and values. Currently, only three attributes are supported:
	// ( About credential request, there are java or golang based implementations:
	//      golang: https://github.com/hyperledger/fabric-ca/blob/fc84b4f088e4a253da012276a4d0b9e3518a3565/lib/client.go#L523
	//  )
	// OU - organization unit of the identity. The value of this attribute is set to identity’s affiliation. For example, if identity’s affiliation is dept1.unit1, then OU attribute is set to dept1.unit1
	// IsAdmin - if the identity is an admin or not. The value of this attribute is set to the value of isAdmin registration attribute.
	// EnrollmentID - enrollment ID of the identity
	/**
	 * TODO WIP
	 * @param {User} admin
	 */
	async idemixEnroll(admin) {
		const client = this._fabricCAClient;
		const {result, errors, messages, success} = await client.post('idemix/credential', {}, admin._signingIdentity);
		const {Nonce, CAInfo: {IssuerPublicKey, IssuerRevocationPublicKey}} = result;
		const nonce = Buffer.from(Nonce, 'base64');
		await client.post('idemix/credential',);
		return result;

	}
}

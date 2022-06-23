import fs from 'fs';
import FormData from 'form-data';
import {RemoteOptsTransform} from '@hyperledger-twgc/fabric-formatter/remote.js';

import {DeliverResponseStatus, DeliverResponseType} from '@hyperledger-twgc/fabric-formatter/eventHub.js';
import EndPoint from 'fabric-common/lib/Endpoint.js';
import Committer from 'fabric-common/lib/Committer.js';
import Eventer from 'fabric-common/lib/Eventer.js';
const {SUCCESS} = DeliverResponseStatus;
const {FULL_BLOCK, STATUS} = DeliverResponseType;

export default class Orderer {
	/**
	 * @param {intString|integer} ordererPort
	 * @param {string} [tlsCaCert] TLS CA certificate file path
	 * @param {CertificatePem} [pem] TLS CA certificate
	 * @param {SSLTargetNameOverride} [ordererHostName]
	 * @param {string} [host]
	 * @param {ClientKey} [clientKey]
	 * @param {ClientCert} [clientCert]
	 * @param {Committer} [committer]
	 * @param logger
	 */
	constructor({ordererPort, tlsCaCert, pem, ordererHostName, host, clientKey, clientCert} = {}, committer, logger = console) {
		if (!committer) {
			if (!pem) {
				if (fs.existsSync(tlsCaCert)) {
					pem = fs.readFileSync(tlsCaCert).toString();
					this.tlsCaCert = tlsCaCert;
				}
			}

			this.host = host ? host : (ordererHostName ? ordererHostName : 'localhost');
			let ordererUrl;
			if (pem) {
				// tls enabled
				ordererUrl = `grpcs://${this.host}:${ordererPort}`;
				this.pem = pem;
				this.sslTargetNameOverride = ordererHostName;
				this.clientKey = clientKey;
				this.clientCert = clientCert;
			} else {
				// tls disabled
				ordererUrl = `grpc://${this.host}:${ordererPort}`;
			}
			const options = RemoteOptsTransform({
				url: ordererUrl,
				host: this.host,
				pem,
				sslTargetNameOverride: this.sslTargetNameOverride,
				clientKey: this.clientKey && fs.readFileSync(this.clientKey).toString(),
				clientCert: this.clientCert && fs.readFileSync(this.clientCert).toString()
			});
			const endpoint = new EndPoint(options);
			committer = new Committer(endpoint.url, null, undefined);
			committer.setEndpoint(endpoint);
		}

		this.committer = committer;

		const {endpoint, serviceClass} = committer;
		const eventer = new Eventer(endpoint.url, {}, undefined);
		eventer.serviceClass = serviceClass;
		eventer.setEndpoint(endpoint);
		this.eventer = eventer;
		this.logger = logger;
	}

	getServiceEndpoints() {
		return [this.committer, this.eventer];
	}

	reset() {
		for (const endpoint of this.getServiceEndpoints()) {
			endpoint.connectAttempted = false;
		}

	}

	async connect() {
		const {logger} = this;
		for (const endpoint of this.getServiceEndpoints()) {
			if (endpoint.connected || endpoint.service) {
				logger.info(`Orderer as [${endpoint.name}] connection exist already`);
			} else {
				await endpoint.connect();
			}
		}

	}

	disconnect() {
		for (const endpoint of this.getServiceEndpoints()) {
			endpoint.disconnect();
		}
	}

	/**
	 * basic health check for an orderer
	 */
	async ping() {
		const {committer} = this;
		try {
			committer.service = new committer.serviceClass(committer.endpoint.addr, committer.endpoint.creds, committer.options);
			await committer.waitForReady(committer.service);
			return true;
		} catch (err) {
			if (err.message.includes('Failed to connect before the deadline')) {
				return false;
			} else {
				throw err;
			}
		}
	}


	/**
	 * Send a Deliver message to the orderer service.
	 *
	 * @param {{signature:Buffer, payload:Buffer}} envelope
	 * @param [requestTimeout]
	 * @returns {Promise<Block[]>}
	 */
	async sendDeliver(envelope, requestTimeout) {
		const {logger} = this;
		const loggerPrefix = `${this.committer.name} sendDeliver`;
		requestTimeout = requestTimeout || this.committer.options.requestTimeout;


		// Send the seek info to the orderer via grpc
		return new Promise((resolve, reject) => {
			const stream = this.committer.service.deliver();
			const responses = [];
			let error_msg = 'SYSTEM_TIMEOUT';

			const deliver_timeout = setTimeout(() => {
				logger.debug(loggerPrefix, `timed out after:${requestTimeout}`);
				stream.end();
				return reject(new Error(error_msg));
			}, requestTimeout);
			stream.on('data', (response) => {
				// DeliverFiltered, DeliverWithPrivateData is designed for peer only
				switch (response.Type) {
					case FULL_BLOCK: {
						const {block} = response;

						logger.debug(loggerPrefix, `received block[${block.header.number}]`);
						responses.push(block);
					}
						break;
					case STATUS: {
						const {status} = response;
						if (status === SUCCESS || status === 200) {
							stream.end();
							return resolve(responses);
						} else {
							stream.end();
							logger.error(loggerPrefix, `rejecting - status:${response.status}`);
							const err = Object.assign(Error('Invalid status returned'), response);
							return reject(err);
						}
					}
					default:
						logger.error(loggerPrefix, `assertion ERROR - unimplemented response.Type=[${response.Type}]`);
						stream.end();
						return reject(new Error('SYSTEM_ERROR'));
				}
			});

			stream.on('status', (response) => {
				logger.debug(loggerPrefix, 'on status', response);
			});

			stream.on('end', () => {
				clearTimeout(deliver_timeout);
				stream.cancel();
				logger.debug(loggerPrefix, 'on end');

			});

			stream.on('error', (err) => {
				logger.debug(loggerPrefix, 'on error');
				stream.end();
				if (err.code === 14) {
					err.originalMessage = err.message;
					err.message = 'SERVICE_UNAVAILABLE';
				}
				return reject(err);
			});

			stream.write(envelope);
			error_msg = 'REQUEST_TIMEOUT';
		});
	}

	toString() {
		return JSON.stringify({Orderer: this.committer.endpoint.url});
	}

	static async join(baseURL, channelName, blockFile, httpClient, adminTLS) {

		const httpOpts = {};
		if (adminTLS) {
			const {clientKey, clientCert, tlsCaCert} = adminTLS;
			httpOpts.key = clientKey; // client key path
			httpOpts.cert = clientCert; // client cert path
			httpOpts.ca = tlsCaCert; // rootCa cert path
		}


		const formData = new FormData();
		formData.append('config-block', fs.createReadStream(blockFile), `${channelName}.block`);

		const url = `${adminTLS ? 'https://' : 'http://'}${baseURL}/participation/v1/channels`;

		try {
			return await httpClient({
				url,
				formData,
				method: 'POST'
			}, httpOpts);
		} catch (e) {
			const {statusCode, statusMessage} = e;
			if (statusCode === 405 && statusMessage === 'Method Not Allowed') {
				const {response: {data: {error}}} = e;
				if (error === 'cannot join: channel already exists') {
					return error;
				}

			}
			throw e;

		}
	}
}

import EndPoint from 'fabric-common/lib/Endpoint.js';
import Endorser from 'fabric-common/lib/Endorser.js';
import Eventer from 'fabric-common/lib/Eventer.js';
import Discoverer from 'fabric-common/lib/Discoverer.js';
import {RemoteOptsTransform} from 'khala-fabric-formatter/remote.js';
import fs from 'fs';
import {BlockEventFilterType} from 'khala-fabric-formatter/eventHub.js';

const {FILTERED_BLOCK, FULL_BLOCK, PRIVATE_BLOCK} = BlockEventFilterType;
export default class Peer {
	/**
	 * @param {intString} peerPort
	 * @param {SSLTargetNameOverride} [peerHostName]
	 * @param {string} [cert] TLS CA certificate file path
	 * @param {CertificatePem} [pem] TLS CA certificate
	 * @param {string} [host]
	 * @param {ClientKey} [clientKey]
	 * @param {ClientCert} [clientCert]
	 * @param {MspId} [mspid]
	 * @param [logger]
	 */
	constructor({peerPort, peerHostName, cert, pem, host, clientKey, clientCert, mspid}, logger = console) {
		this.logger = logger;
		if (!pem) {
			if (fs.existsSync(cert)) {
				pem = fs.readFileSync(cert).toString();
			}
		}

		this.host = host ? host : (peerHostName ? peerHostName : 'localhost');
		let peerUrl;
		if (pem) {
			// tls enabled
			peerUrl = `grpcs://${this.host}:${peerPort}`;
			this.pem = pem;
			this.sslTargetNameOverride = peerHostName;
			this.clientKey = clientKey;
			this.clientCert = clientCert;
		} else {
			// tls disabled
			peerUrl = `grpc://${this.host}:${peerPort}`;
		}


		const options = RemoteOptsTransform({
			url: peerUrl,
			host: this.host,
			pem,
			sslTargetNameOverride: this.sslTargetNameOverride,
			clientKey: this.clientKey && fs.readFileSync(this.clientKey).toString(),
			clientCert: this.clientCert && fs.readFileSync(this.clientCert).toString()
		});
		const endpoint = new EndPoint(options);
		const endorser = new Endorser(endpoint.url, {}, mspid);
		endorser.setEndpoint(endpoint);
		this.endorser = endorser;

		const eventer = new Eventer(endpoint.url, {}, mspid);
		eventer.setEndpoint(endpoint);
		this.eventer = eventer;

		const discoverer = new Discoverer(endpoint.url, {}, mspid);
		discoverer.setEndpoint(endpoint);
		this.discoverer = discoverer;
	}

	get mspId() {
		return this.endorser.mspid;
	}

	getServiceEndpoints() {
		return [this.endorser, this.eventer, this.discoverer];
	}

	async connect() {
		const {logger} = this;
		for (const serviceEndpoint of this.getServiceEndpoints()) {
			if (serviceEndpoint.connected || serviceEndpoint.service) {
				logger.info(`${serviceEndpoint.type} ${serviceEndpoint.name} connection exist already`);
			} else {
				await serviceEndpoint.connect();
			}
		}
	}

	disconnect() {
		this.getServiceEndpoints().forEach((serviceEndpoint) => {
			serviceEndpoint.disconnect();
		});
	}

	/**
	 * basic health check as endorser role
	 * @return {Promise<boolean>} false if connect trial failed
	 */
	async ping() {
		try {
			const {endorser} = this;
			const {endpoint} = endorser;
			endorser.service = new endorser.serviceClass(endpoint.addr, endpoint.creds, endorser.options);
			await endorser.waitForReady(endorser.service);
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
	 * TODO WIP stream.resume();
	 * check if the stream is ready.
	 * The stream must be readable, writeable and reading to be 'ready' and not paused.
	 */
	isStreamReady() {
		const {eventer: {stream}} = this;

		return !!stream && !stream.isPaused() && stream.readable && stream.writable;

	}

	/**
	 * get a new stream based on block type
	 */
	async connectStream(blockType) {
		const {eventer} = this;

		eventer.service = new eventer.serviceClass(eventer.endpoint.addr, eventer.endpoint.creds, eventer.options);
		await eventer.waitForReady(eventer.service);

		switch (blockType) {
			case FILTERED_BLOCK:
				eventer.stream = eventer.service.deliverFiltered();
				break;
			case FULL_BLOCK:
				eventer.stream = eventer.service.deliver();
				break;
			case PRIVATE_BLOCK:
				eventer.stream = eventer.service.deliverWithPrivateData();
				break;
			default:
				eventer.stream = eventer.service.deliver();
		}
	}

	async disconnectStream() {
		const {eventer} = this;

		return new Promise((resolve, reject) => {
			eventer.stream.once('error', e => {
				const {code, details} = e;
				if (code === 1 && details === 'Cancelled on client') {
					resolve(details);
				} else {
					reject(e);
				}
			});

			eventer.stream.cancel();
			eventer.stream.end();
			delete eventer.stream;
		});

	}

	toString() {
		return JSON.stringify({Peer: this.endorser.endpoint.url});
	}
}

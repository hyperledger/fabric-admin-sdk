import DiscoveryService from 'fabric-common/lib/DiscoveryService.js';
import {emptyChannel} from './channel.js';

export default class SlimDiscoveryService extends DiscoveryService {

	constructor(channelName, discoverer) {
		super('-', emptyChannel(channelName));
		delete this.refreshAge;

		delete this.discoveryResults;
		delete this.asLocalhost;
		delete this.targets;
		delete this.setTargets;

		this.setDiscoverer(discoverer);
	}

	setDiscoverer(discoverer) {
		this.currentTarget = discoverer;
	}

	/**
	 * @typedef {Object} DiscoveryResult
	 * @property {[]} results
	 * @property {Discoverer} connection
	 * @property {string} peer
	 */
	/**
	 *
	 * @param request
	 * @return {Promise<DiscoveryResult>}
	 */
	async send(request = {}) {
		const {requestTimeout = this.requestTimeout, target = this.currentTarget} = request;

		const signedEnvelope = this.getSignedEnvelope();
		const response = await target.sendDiscovery(signedEnvelope, requestTimeout);
		if (response instanceof Error) {
			throw response;
		}
		return response;
	}

	build(idContext, {config = null, local = null, interest, onlineSign = true}) {
		if (config) {
			local = false; // otherwise we will have multiple result with type 'members'
		}
		const result = super.build(idContext, {config, local, interest, endorsement: null});// endorsement work as a helper to build interest
		if (onlineSign) {
			super.sign(idContext);
		}
		return result;
	}

}

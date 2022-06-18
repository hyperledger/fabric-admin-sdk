import fs from 'fs';
import {OrdererType} from './constants.js';

export default class ConfigFactory {
	/**
	 * @param {string|json} original_config
	 * @param logger
	 */
	constructor(original_config, logger = console) {
		this.newConfig = JSON.parse(original_config);
		this.logger = logger;
	}

	static _toBase64(pem) {
		return fs.readFileSync(pem).toString('base64');
	}

	/**
	 * @param {OrgName} OrgName
	 * @param {ChannelGroupType} channelGroupType
	 * @return {ConfigFactory}
	 */
	deleteMSP(OrgName, channelGroupType) {
		delete this.newConfig.channel_group.groups[channelGroupType].groups[OrgName];
		return this;
	}

	/**
	 *
	 * @param {MspId} MSPID
	 * @param {ChannelGroupType} channelGroupType
	 * @return {ConfigFactory}
	 */
	assignDictator(MSPID, channelGroupType) {
		const target = channelGroupType;
		this.newConfig.channel_group.groups[target].policies.Admins.policy = {
			type: 1,
			value: {
				identities: [
					{
						principal: {
							msp_identifier: MSPID,
							role: 'ADMIN'
						},
						principal_classification: 'ROLE'
					}
				],
				rule: {
					n_out_of: {
						n: 1,
						rules: [
							{
								signed_by: 0
							}
						]
					}
				}
			}
		};
		return this;
	}

	/**
	 *
	 * @param OrgName
	 * @param {ChannelGroupType} channelGroupType
	 */
	getOrg(OrgName, channelGroupType) {
		return this.newConfig.channel_group.groups[channelGroupType].groups[OrgName];
	}

	/**
	 * @param {string} OrgName
	 * @param {ChannelGroupType} channelGroupType
	 * @param admin
	 * @return {ConfigFactory}
	 */
	addAdmin(OrgName, channelGroupType, admin) {
		if (!this.getOrg(OrgName, channelGroupType)) {
			this.logger.error(OrgName, 'not exist, addAdmin skipped');
			return this;
		}
		const adminCert = fs.readFileSync(admin).toString('base64');
		const admins = this.newConfig.channel_group.groups[channelGroupType].groups[OrgName].values.MSP.value.config.admins;
		if (admins.find(adminCert)) {
			this.logger.warn('adminCert found, addAdmin skipped');
			return this;
		}
		admins.push(adminCert);
		return this;
	}

	/**
	 *
	 * @param {OrgName} orgName
	 */
	getAnchorPeers(orgName) {
		return this.newConfig.channel_group.groups.Application.groups[orgName].values.AnchorPeers;
	}

	/**
	 * @param {OrgName} orgName
	 * @param {[{host:string,port:number}]} anchorPeers
	 */
	setAnchorPeers(orgName, anchorPeers) {
		anchorPeers = anchorPeers.map(({host, port}) => ({host: host.toString(), port}));
		const {AnchorPeers} = this.newConfig.channel_group.groups.Application.groups[orgName].values;
		if (AnchorPeers) {
			AnchorPeers.value.anchor_peers = anchorPeers;
		} else {
			this.newConfig.channel_group.groups.Application.groups[orgName].values.AnchorPeers = {
				mod_policy: 'Admins',
				value: {
					anchor_peers: anchorPeers
				},
				version: '0'
			};
		}
		return this;
	}

	/**
	 * because we will not change the 'version' property, so it will never be totally identical
	 * @param {OrgName} OrgName
	 * @param {MspId} MSPID
	 * @param {ChannelGroupType} channelGroupType
	 * @param {string[]} admins pem file path array
	 * @param {string[]} root_certs pem file path array
	 * @param {string[]} tls_root_certs pem file path array
	 * @param skipIfExist
	 * @return {ConfigFactory}
	 */
	createOrUpdateOrg(OrgName, MSPID, channelGroupType, {admins = [], root_certs = [], tls_root_certs = []} = {}, skipIfExist) {
		if (skipIfExist && this.getOrg(OrgName, channelGroupType)) {
			this.logger.info(OrgName, 'exist, createOrUpdateOrg skipped');
			return this;
		}
		const target = channelGroupType;
		this.newConfig.channel_group.groups[target].groups[OrgName] = {
			mod_policy: 'Admins',
			policies: {
				Admins: {
					mod_policy: 'Admins',
					policy: {
						type: 1,
						value: {
							identities: [
								{
									principal: {
										msp_identifier: MSPID,
										role: 'ADMIN'
									},
									principal_classification: 'ROLE'
								}
							],
							rule: {
								n_out_of: {
									n: 1,
									rules: [
										{
											signed_by: 0
										}
									]
								}
							}
						}
					}
				},
				Readers: {
					mod_policy: 'Admins',
					policy: {
						type: 1,
						value: {
							identities: [
								{
									principal: {
										msp_identifier: MSPID,
										role: 'MEMBER'
									},
									principal_classification: 'ROLE'
								}
							],
							rule: {
								n_out_of: {
									n: 1,
									rules: [
										{
											signed_by: 0
										}
									]
								}
							}
						}
					}
				},
				Writers: {
					mod_policy: 'Admins',
					policy: {
						type: 1,
						value: {
							identities: [
								{
									principal: {
										msp_identifier: MSPID,
										role: 'MEMBER'
									},
									principal_classification: 'ROLE'
								}
							],
							rule: {
								n_out_of: {
									n: 1,
									rules: [
										{
											signed_by: 0
										}
									]
								}
							}
						}
					}
				}
			},
			values: {
				MSP: {
					mod_policy: 'Admins',
					value: {
						config: {
							admins: admins.map(ConfigFactory._toBase64),
							crypto_config: {
								identity_identifier_hash_function: 'SHA256',
								signature_hash_family: 'SHA2'
							},
							name: MSPID,
							root_certs: root_certs.map(ConfigFactory._toBase64),
							tls_root_certs: tls_root_certs.map(ConfigFactory._toBase64)
						},
						type: 0
					}
				}
			}

		};
		return this;
	}

	/**
	 *
	 * @param {string} version
	 * @return {string}
	 * @private
	 */
	static _fromSemVer(version) {
		return `V${version.replace('.', '_')}`;
	}

	setCapability(version) {
		const {Capabilities} = this.newConfig.channel_group.groups.Orderer.values;
		if (!Capabilities) {
			this.newConfig.channel_group.groups.Orderer.values.Capabilities = {
				mod_policy: 'Admins',
				value: {
					capabilities: {
						[ConfigFactory._fromSemVer(version)]: {}
					}
				}
			};
		} else {
			Capabilities.value.capabilities = {
				[ConfigFactory._fromSemVer(version)]: {}
			};
		}
		return this;
	}

	/**
	 * @param newAddr
	 * @return {ConfigFactory}
	 */
	addOrdererAddress(newAddr) {
		if (!this.newConfig.channel_group.values.OrdererAddresses.value.addresses.includes(newAddr)) {
			this.newConfig.channel_group.values.OrdererAddresses.value.addresses.push(newAddr);
		} else {
			this.logger.info(newAddr, 'exist, addOrdererAddress skipped');
		}
		return this;
	}

	getOrdererAddresses() {
		return this.newConfig.channel_group.values.OrdererAddresses.value.addresses;
	}

	/**
	 *
	 * @param {OrdererType} type
	 */
	setConsensusType(type) {
		switch (type) {
			case OrdererType.etcdraft:
				this.newConfig.channel_group.groups.Orderer.values.ConsensusType.value.type = type;
				delete this.newConfig.channel_group.groups.Orderer.values.KafkaBrokers;
				break;
			default:
				throw Error(`Unsupported ConsensusType ${type}`);
		}
		return this;
	}

	setConsensusMetadata(consenters, options) {
		const {metadata} = this.newConfig.channel_group.groups.Orderer.values.ConsensusType.value;
		if (metadata) {
			if (options) {
				metadata.options = options;
			}
			if (consenters) {
				metadata.consenters = consenters;
			}
		} else {
			if (!options) {
				options = {
					election_tick: 10,
					heartbeat_tick: 1,
					max_inflight_blocks: 5,
					snapshot_interval_size: 20971520,
					tick_interval: '500ms'
				};
			}
			this.newConfig.channel_group.groups.Orderer.values.ConsensusType.value.metadata = {
				consenters: consenters.map(({client_tls_cert, host, port = 7050, server_tls_cert}) => ({
					client_tls_cert: ConfigFactory._toBase64(client_tls_cert),
					server_tls_cert: ConfigFactory._toBase64(server_tls_cert),
					host, port: parseInt(port)
				})),
				options
			};
		}
		return this;
	}

	/**
	 * setting the ordering service into maintenance mode
	 * @param isDirectionIn true to setting the ordering service into maintenance mode, false to back to normal mode
	 * @return {ConfigFactory}
	 */
	maintenanceMode(isDirectionIn) {
		this.newConfig.channel_group.groups.Orderer.values.ConsensusType.value.state = isDirectionIn ? 'STATE_MAINTENANCE' : 'STATE_NORMAL';
		return this;
	}

	build() {
		return JSON.stringify(this.newConfig);
	}
}

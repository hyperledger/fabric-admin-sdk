import Channel from 'fabric-common/lib/Channel.js';

export const emptyChannel = (channelName) => {
	const client = {
		getClientCertHash: () => Buffer.from(''),
		getConfigSetting: () => undefined
	};
	return new Channel(channelName, client);
};

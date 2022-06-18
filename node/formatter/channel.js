/**
 * @param {string} channelName
 * @param {boolean} [toThrow]
 */
export const nameMatcher = (channelName, toThrow) => {
	const namePattern = /^[a-z][a-z0-9.-]*$/;
	const result = channelName.match(namePattern) && channelName.length < 250;
	if (!result && toThrow) {
		throw Error(`invalid channel name ${channelName}; should match regx: ${namePattern} and with length < 250`);
	}
	return result;
};

export const configGroupMatcher = (configGroupName, toThrow) => {
	const namePattern = /^[a-zA-Z0-9.-]+$/;
	const result = configGroupName.match(namePattern) && configGroupName.length < 250;
	if (!result && toThrow) {
		throw Error(`invalid config group name [${configGroupName}] should match regx: ${namePattern} and with length < 250`);
	}
	return result;
};

/**
 *
 * @param {MspId} mspid
 * @return {boolean}
 */
export const mspIdMatcher = (mspid) => {
	const namePattern = /^[a-zA-Z0-9.-]+$/;
	return !!mspid.match(namePattern);
};

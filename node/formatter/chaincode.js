export const parsePackageID = (packageId) => {
	const [name, hash] = packageId.split(':');
	return {name, hash, label: name};
};
export const nameMatcher = (chaincodeName) => {
	const namePattern = /^[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*$/;
	return chaincodeName.match(namePattern);
};
export const versionMatcher = (ccVersionName) => {
	const namePattern = /^[A-Za-z0-9_.+-]+$/;
	return ccVersionName.match(namePattern);
};

export const collectionMatcher = (collectionName) => {
	const namePattern = /^[A-Za-z0-9-]+([A-Za-z0-9_-]+)*$/;
	return collectionName.match(namePattern);
};
export const packageFileMatcher = (packageFileName) => {
	const namePattern = /^(.+)[.]([0-9a-f]{64})[.]tar[.]gz$/;
	return packageFileName.match(namePattern);
};

/**
 * @enum
 */
export const ChaincodeSpecType = {
	UNDEFINED: 0,
	GOLANG: 1,
	NODE: 2,
	CAR: 3,
	JAVA: 4,
};

/**
 * @enum {string}
 */
export const ChaincodeType = {
	golang: 'golang',
	node: 'node',
	java: 'java',
	external: 'ccaas',
};
export const implicitCollection = (mspid) => `_implicit_org_${mspid}`;

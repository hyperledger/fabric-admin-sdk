export const toString = (caService) => {
	const caClient = caService._fabricCAClient;
	const returned = {
		caName: caClient._caName,
		hostname: caClient._hostname,
		port: caClient._port
	};
	const trustedRoots = caClient._tlsOptions.trustedRoots.map(buffer => buffer.toString());
	returned.tlsOptions = {
		trustedRoots,
		verify: caClient._tlsOptions.verify
	};

	return JSON.stringify(returned);
};
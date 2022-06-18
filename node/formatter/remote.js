/**
 * @typedef {string} ClientKey The private key file, in PEM format
 *    To use with the gRPC protocol (that is, with TransportCredentials).
 *    Required when using the grpcs protocol with client certificates.
 */

/**
 * @typedef {string} ClientCert The public certificate file, in PEM format,
 *    To use with the gRPC protocol (that is, with TransportCredentials).
 *    Required when using the grpcs protocol with client certificates.
 */

/**
 * @deprecated deprecated, replace it with SAN entry DNS=localhost
 * @typedef {string} SSLTargetNameOverride
 * Used in test environment only
 */

export const defaultGRPCOptions = {
	'grpc.max_receive_message_length': -1,
	'grpc.max_send_message_length': -1,
	'grpc.keepalive_time_ms': 120000,
	'grpc.http2.min_time_between_pings_ms': 120000,
	'grpc.keepalive_timeout_ms': 20000,
	'grpc.http2.max_pings_without_data': 0,
	'grpc.keepalive_permit_without_calls': 1,
};

/**
 *
 * @param {Object} opts
 * @param [logger]
 */
export const RemoteOptsTransform = (opts = {}, logger = console) => {
	const {sslTargetNameOverride, host, waitForReadyTimeout, requestTimeout} = opts;

	if (host && host.toLowerCase() !== host) {
		logger.error(`invalid hostname [${host}] : [docker-network][gRPCs] host endpoint contains upper case is not allowed in TLS auth within docker network`);
	}
	if (sslTargetNameOverride) {
		opts['grpc.ssl_target_name_override'] = sslTargetNameOverride;
		logger.warn(`[grpc.ssl_target_name_override]=${sslTargetNameOverride} should be used for test environment only. Please consider add entry DNS=localhost to SAN extension of the server certificate`);
		opts['grpc.default_authority'] = sslTargetNameOverride;
		delete opts.sslTargetNameOverride;
	}
	opts['grpc-wait-for-ready-timeout'] = waitForReadyTimeout && Number.isInteger(waitForReadyTimeout) ? waitForReadyTimeout : 3000;// default 3 seconds
	delete opts.waitForReadyTimeout;

	opts.requestTimeout = requestTimeout && Number.isInteger(requestTimeout) ? requestTimeout : 3000; // default 3 seconds

	for (const [key, value] of Object.entries(opts)) {
		if (!value) {
			delete opts[key];
		}
	}
	return Object.assign({}, defaultGRPCOptions, opts);
};


/**
 * Valid logging levels are case-insensitive string
 * @enum {string}
 */
export const LoggingLevel = {
	fatal: 'FATAL', FATAL: 'FATAL',
	panic: 'PANIC', PANIC: 'PANIC',
	error: 'ERROR', ERROR: 'ERROR',
	warning: 'WARNING', WARNING: 'WARNING',
	info: 'INFO', INFO: 'INFO',
	debug: 'DEBUG', DEBUG: 'DEBUG',
	undefined: 'INFO', null: 'INFO'
};
export const rootCAsStringBuilder = ({caCert, rootCAs}) => {
	let result = [caCert];
	if (Array.isArray(rootCAs)) {
		result = result.concat(rootCAs);
	}
	return result.join(',');
};

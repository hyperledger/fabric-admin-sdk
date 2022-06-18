import Utils from 'fabric-common/lib/Utils.js';

/**
 *
 * @return {ICryptoSuite}
 */
export const emptySuite = () => Utils.newCryptoSuite();

export const HSMSuite = ({lib, slot, pin}) => {
	return Utils.newCryptoSuite({software: false, lib, slot, pin, keysize: 256, hash: 'SHA2'}); // software false to use HSM
};

import path from 'path';
import fs from 'fs';

export class ECDSA_PrvKey {
	/**
	 * @param key
	 */
	constructor(key) {
		this._key = key;
	}

	pem() {
		return this._key.toBytes();
	}

	/**
	 * @deprecated use static filename "priv_sk" from 2.x
	 * fabric private key raw PEM filename
	 * @return {string}
	 */
	filename() {
		const {prvKeyHex} = this._key;
		if (!prvKeyHex) {
			throw Error('not private key');
		}
		return `${prvKeyHex}_sk`;
	}

	toKeystore(dirName) {
		const filename = 'priv_sk';
		const absolutePath = path.resolve(dirName, filename);
		const data = this.pem();
		fs.mkdirSync(dirName, {recursive: true});
		fs.writeFileSync(absolutePath, data);

	}
}

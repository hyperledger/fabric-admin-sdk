import fs from 'fs';
import path from 'path';
export const findKeyFiles = (dir) => {
	const files = fs.readdirSync(dir);
	return files.filter((fileName) => fileName.endsWith('_sk')).map((fileName) => path.resolve(dir, fileName));
};
export const getOneKeystore = (dir) => {
	const files = fs.readdirSync(dir);
	const filename = files.find((fileName) => fileName.endsWith('_sk'));
	if (filename) {
		return fs.readFileSync(path.resolve(dir, filename)).toString();
	}
};
/*
 * Make sure there's a start line with '-----BEGIN CERTIFICATE-----'
 * and end line with '-----END CERTIFICATE-----', so as to be compliant
 * with x509 parsers
 */
export const normalizeX509 = (raw) => {
	const regex = /(-----\s*BEGIN ?[^-]+?-----)([\s\S]*)(-----\s*END ?[^-]+?-----)/;
	let matches = raw.match(regex);
	if (!matches || matches.length !== 4) {
		throw new Error('Failed to find start line or end line of the certificate.');
	}

	// remove the first element that is the whole match
	matches.shift();
	// remove LF or CR
	matches = matches.map((element) => {
		return element.trim();
	});

	// make sure '-----BEGIN CERTIFICATE-----' and '-----END CERTIFICATE-----' are in their own lines
	// and that it ends in a new line
	let result = matches.join('\n') + '\n';
	// could be this has multiple certs within that are not separated by a newline
	const regex2 = /----------/;
	result = result.replace(new RegExp(regex2, 'g'), '-----\n-----');
	return result;
};
export const findCertFiles = (dir) => {
	const files = fs.readdirSync(dir);
	return files.map((fileName) => path.resolve(dir, fileName)).filter(filePath => {
		try {
			const pem = fs.readFileSync(filePath).toString();
			normalizeX509(pem);
			return true;
		} catch (e) {
			return false;
		}
	});
};

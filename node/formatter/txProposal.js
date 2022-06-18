/**
 *
 * @param {TransientMap} jsObject
 * @return {Client.TransientMap}
 */
export const transientMapTransform = (jsObject) => {
	if (!jsObject) {
		return null;
	}
	const result = {};
	for (const [key, value] of Object.entries(jsObject)) {
		result[key] = Buffer.from(value);
	}
	return result;
};

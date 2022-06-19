/**
 *
 * @param {Message|Object} protobufMessage
 * @param {constructor} [asType] required if protobufMessage is a js object. The class of protobuf.
 * @return {Buffer}
 */
export const BufferFrom = (protobufMessage, asType) => {
	let message
	if (asType) {
		message = ProtoFrom(protobufMessage, asType)
	} else {
		message = protobufMessage
	}
	return Buffer.from(message.serializeBinary())
}
/**
 *
 * @param {Object} object
 * @param {constructor} asType The class of protobuf.
 * @return {Message}
 */
export const ProtoFrom = (object, asType) => {
	const message = new asType()
	for (const [key, value] of Object.entries(object)) {
		let fcn = `set${key[0].toUpperCase() + key.slice(1)}`
		if (Array.isArray(value)) {
			fcn += 'List'
		}
		if (typeof message[fcn] !== "function") {
			console.debug(fcn)
		}
		message[fcn](value)
	}
	return message;
}

/**
 *
 * @param {Uint8Array} binary
 * @param {constructor} asType The class of protobuf.
 * @return {Object}
 */
export const decode = (binary, asType)=>{
	const message = asType.deserializeBinary(binary);
	return message.toObject()
}
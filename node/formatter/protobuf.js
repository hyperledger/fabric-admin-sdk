export const BufferFrom = (protobufMessage, asType = protobufMessage.constructor) => asType.encode(protobufMessage).finish();

export const ProtoFrom = (object, asType) => asType.fromObject(object);

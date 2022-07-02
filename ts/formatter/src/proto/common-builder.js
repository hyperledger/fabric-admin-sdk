import { Timestamp } from 'google-protobuf/google/protobuf/timestamp_pb';
export class Builder {
	static currentTimestamp () {
		const now = new Date();
		const timestamp = new Timestamp();
		timestamp.setSeconds(now.getTime() / 1000);
		timestamp.setNanos((now.getTime() % 1000) * 1000000)
		return timestamp;
	}
}

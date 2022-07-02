import {msp} from '@hyperledger/fabric-protos'
const {SerializedIdentity} = msp
describe('syntax tests', function(){
	this.timeout(0)
	it('toBuffer method', ()=>{
		const proto = new SerializedIdentity()
		const bytes = proto.serializeBinary()

	})
})
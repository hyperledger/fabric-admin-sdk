import {HeaderType, Status} from '../lib/constants'
import {CommitResponse} from '../lib/chaincode'

describe('ts syntax', function () {
    it('ValueOf type', () => {
        const resp: CommitResponse = {
            info: '123',
            status: Status.SUCCESS
        }
    })
    it('enum: getName', () => {
        expect(HeaderType[0]).toBe('MESSAGE')
    })
});
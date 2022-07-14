import {Name} from '../lib/names'

describe('name match tests', function () {

    it('couch db name', () => {
        expect(new Name('abc').isCouchDB()).toBe(true)
        expect(new Name('123').isCouchDB()).toBe(false)
    })

    describe('@deprecated', ()=>{
        it('chaincode name/version', () => {
            expect(new Name('_test').isChaincode()).toBe(false)
            expect(new Name('123').isChaincode()).toBe(true)

            expect(new Name('123').isChaincodeVersion()).toBe(true)
            expect(new Name('1.1.*').isChaincodeVersion()).toBe(false)
        })
    })
    it('collection name', () => {
        expect(new Name('123').isCollection()).toBe(true)
        expect(new Name('_implict').isCollection()).toBe(false)
    })

    it('package file', () => {
        const file = 'diagnosis.b54e2b152c2135f45dfca42c67ec11c8a8d6076f1796f3deaa64a752c1d4efa8.tar.gz'
        expect(new Name(file).isPackageFile()).toBe(true)
        expect(new Name('_temp').isPackageFile()).toBe(false)
    })


})

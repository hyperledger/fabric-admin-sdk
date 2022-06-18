import GatePolicy from '../gatePolicy.js';
import {FromStandard} from '../SideDB.js';
import {consoleLogger} from '@davidkhala/logger/log4.js';

const logger = consoleLogger('test:gate policy');
describe('policy parser', () => {

	it('orOfAnds', () => {
		const policyStr = `OR(AND('A.member', 'B.member'), 'C.member', AND('A.member', 'D.member'))`;
		GatePolicy.FromString(policyStr);
	});
	it('andOfOrs', () => {
		const policyStr = `AND('A.member', 'C.member', OR('B.member', 'D.member'))`;
		GatePolicy.FromString(policyStr);
	});
	it('orOfOrs', () => {
		const policyStr = `OR('A.member', OR('B.member', 'C.member'))`;
		GatePolicy.FromString(policyStr);
	});
	it('andOfAnds', () => {
		const policyStr = `AND('A.member', AND('B.member', 'C.member'), AND('D.member','A.member'))`;
		GatePolicy.FromString(policyStr);
	});
	it('RoleClausePattern', () => {
		const str = `'abc.org.member'`;
		const result = str.match(GatePolicy.RoleClausePattern);
		logger.info(result);
	});

});
describe('standard "collections_config.json" translator', () => {
	const json = [
		{
			'name': 'collectionMarbles',
			'policy': 'OR(\'Org1MSP.member\', \'Org2MSP.member\')',
			'requiredPeerCount': 0,
			'maxPeerCount': 3,
			'blockToLive': 1000000,
			'memberOnlyRead': true
		},

		{
			'name': 'collectionMarblePrivateDetails',
			'policy': 'OR(\'Org1MSP.member\')',
			'requiredPeerCount': 0,
			'maxPeerCount': 3,
			'blockToLive': 3,
			'memberOnlyRead': true
		}
	];
	const results = FromStandard(json);
	logger.info(results);
});

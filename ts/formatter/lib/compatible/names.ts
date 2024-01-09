export class Name extends String {


    /**
     * @deprecated stay in sync with those defined in core/scc/lscc/lscc.go
     */
    isChaincode() {
        const namePattern = /^[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*$/;
        return !!this.match(namePattern);
    }

    /**
     * @deprecated stay in sync with those defined in core/scc/lscc/lscc.go
     */
    isChaincodeVersion() {
        const namePattern = /^[A-Za-z0-9_.+-]+$/;
        return !!this.match(namePattern);
    }
}


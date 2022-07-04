export class Name extends String {
    isCouchDB(): boolean {
        const expectedDatabaseNamePattern = /^[a-z][a-z0-9.$_()+-]*$/;
        return !!this.match(expectedDatabaseNamePattern)
    }

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

    isCollection() {
        const namePattern = /^[A-Za-z0-9-]+([A-Za-z0-9_-]+)*$/;
        return !!this.match(namePattern);
    }

    isPackageFile() {
        const namePattern = /^(.+)[.]([0-9a-f]{64})[.]tar[.]gz$/;
        return !!this.match(namePattern);
    }
}


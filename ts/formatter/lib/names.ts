export class Name extends String {
    isCouchDB(): boolean {
        const expectedDatabaseNamePattern = /^[a-z][a-z0-9.$_()+-]*$/;
        return !!this.match(expectedDatabaseNamePattern)
    }

    isMspId(): boolean {
        const namePattern = /^[a-zA-Z0-9.-]+$/;
        return !!this.match(namePattern);
    }

    isChannel(): boolean {
        const namePattern = /^[a-z][a-z0-9.-]*$/;
        return this.match(namePattern) && this.length < 250;
    }

    isChannelConfigGroup(): boolean {
        const namePattern = /^[a-zA-Z0-9.-]+$/;
        return this.match(namePattern) && this.length < 250;
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


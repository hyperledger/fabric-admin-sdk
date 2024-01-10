export class Name extends String {
    isCouchDB(): boolean {
        const expectedDatabaseNamePattern = /^[a-z][a-z0-9.$_()+-]*$/;
        return this._match(expectedDatabaseNamePattern)
    }

    _match(regex:RegExp){
        return regex.test(this.toString())
    }

    isMspId(): boolean {
        const namePattern = /^[a-zA-Z0-9.-]+$/;
        return this._match(namePattern);
    }

    parsePackageID(packageId: string) {
        const [label, hash] = packageId.split(':');
        return {hash, label};
    }

    isChannel(): boolean {
        const namePattern = /^[a-z][a-z0-9.-]*$/;
        return this._match(namePattern) && this.length < 250;
    }

    isChannelConfigGroup(): boolean {
        const namePattern = /^[a-zA-Z0-9.-]+$/;
        return this._match(namePattern) && this.length < 250;
    }


    isCollection() {
        const namePattern = /^[A-Za-z0-9-]+([A-Za-z0-9_-]+)*$/;
        return this._match(namePattern);
    }

    isPackageFile() {
        const namePattern = /^(.+)[.]([0-9a-f]{64})[.]tar[.]gz$/;
        return this._match(namePattern);
    }
}


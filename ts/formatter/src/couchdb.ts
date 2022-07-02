export class CouchDBName extends String {
    validate(): boolean {
        const expectedDatabaseNamePattern = /^[a-z][a-z0-9.$_()+-]*$/;
        return !!this.match(expectedDatabaseNamePattern)
    }
}


export const DatabaseNameMatcher = (databaseName) => {
	const expectedDatabaseNamePattern = /^[a-z][a-z0-9.$_()+-]*$/;
	return databaseName.match(expectedDatabaseNamePattern);
};

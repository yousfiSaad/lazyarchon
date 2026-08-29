package local

import _ "embed"

// migrationV1 is the initial schema. Migrations are applied in order and
// tracked with PRAGMA user_version: migrations[i] runs when user_version == i,
// and user_version is set to i+1 afterwards. Never edit an applied migration;
// append a new one instead.
//
//go:embed migrations/001_init.sql
var migrationV1 string

// migrations holds one SQL script per schema version, in order.
var migrations = []string{migrationV1}

// schemaVersion returns the current schema version (the number of applied
// migrations once the database is fully migrated).
func schemaVersion() int {
	return len(migrations)
}

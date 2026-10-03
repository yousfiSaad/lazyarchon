// Package local provides a SQLite-backed task backend for lazyarchon.
//
// It stores projects and tasks in a single local database file (WAL mode),
// requires no network or external services, and implements the full
// plugin.TaskClient surface including project write operations. The database
// path defaults to the user's data directory and can be overridden with the
// LAZYARCHON_DB_PATH environment variable or the plugins.local.path config.
package local

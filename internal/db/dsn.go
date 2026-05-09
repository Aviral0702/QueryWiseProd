package db

import (
	"net/url"
	"strings"
)

// DSNLabel extracts a display database name and host:port (or host) from a Postgres DSN URL.
func DSNLabel(dsn string) (databaseName, hostEndpoint string) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "unknown", "unknown"
	}

	db := strings.TrimPrefix(u.Path, "/")
	if db == "" {
		db = "postgres"
	}

	host := u.Host
	if host == "" {
		host = "localhost"
	}

	return db, host
}

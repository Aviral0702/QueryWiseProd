package db

import (
	"fmt"
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

// weakSSLModes are explicit modes that permit unencrypted or unverified connections.
var weakSSLModes = map[string]bool{
	"disable": true,
	"allow":   true,
	"prefer":  true,
}

// ensureSSLMode inspects a DSN's sslmode and enforces a secure-by-default choice.
// If sslmode is absent it forces sslmode=require so the connection is always
// encrypted. If sslmode is explicitly set to a weak mode (disable/allow/prefer)
// the user's choice is kept but a non-empty warning is returned. Stronger modes
// (require/verify-ca/verify-full) are left unchanged with no warning. It supports
// both URL DSNs (postgres://...) and keyword DSNs (host=... sslmode=...).
func ensureSSLMode(dsn string) (updated string, warning string) {
	if strings.Contains(dsn, "://") {
		return ensureSSLModeURL(dsn)
	}
	return ensureSSLModeKeyword(dsn)
}

func ensureSSLModeURL(dsn string) (updated string, warning string) {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn, ""
	}
	q := u.Query()
	mode := strings.ToLower(strings.TrimSpace(q.Get("sslmode")))
	if mode == "" {
		q.Set("sslmode", "require")
		u.RawQuery = q.Encode()
		return u.String(), ""
	}
	if weakSSLModes[mode] {
		return dsn, weakSSLWarning(mode)
	}
	return dsn, ""
}

func ensureSSLModeKeyword(dsn string) (updated string, warning string) {
	mode, found := keywordSSLMode(dsn)
	if !found {
		return dsn + " sslmode=require", ""
	}
	if weakSSLModes[mode] {
		return dsn, weakSSLWarning(mode)
	}
	return dsn, ""
}

// keywordSSLMode does a case-insensitive scan for an sslmode=... keyword and
// returns its lower-cased value.
func keywordSSLMode(dsn string) (mode string, found bool) {
	for _, field := range strings.Fields(dsn) {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), "sslmode") {
			return strings.ToLower(strings.TrimSpace(value)), true
		}
	}
	return "", false
}

func weakSSLWarning(mode string) string {
	return fmt.Sprintf("warning: sslmode=%q permits unencrypted or unverified connections; use sslmode=verify-full with sslrootcert for production", mode)
}

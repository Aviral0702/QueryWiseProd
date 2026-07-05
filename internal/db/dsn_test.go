package db

import "testing"

func TestEnsureSSLMode(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		wantDSN     string
		wantWarning bool
	}{
		{
			name:    "url without sslmode adds require",
			dsn:     "postgres://user:pass@host:5432/db",
			wantDSN: "postgres://user:pass@host:5432/db?sslmode=require",
		},
		{
			name:        "url with sslmode=disable unchanged with warning",
			dsn:         "postgres://user:pass@host:5432/db?sslmode=disable",
			wantDSN:     "postgres://user:pass@host:5432/db?sslmode=disable",
			wantWarning: true,
		},
		{
			name:    "url with verify-full unchanged no warning",
			dsn:     "postgres://user:pass@host:5432/db?sslmode=verify-full",
			wantDSN: "postgres://user:pass@host:5432/db?sslmode=verify-full",
		},
		{
			name:    "keyword form without sslmode appends require",
			dsn:     "host=localhost port=5432 dbname=db user=me",
			wantDSN: "host=localhost port=5432 dbname=db user=me sslmode=require",
		},
		{
			name:    "keyword form with sslmode=require unchanged no warning",
			dsn:     "host=localhost port=5432 dbname=db sslmode=require",
			wantDSN: "host=localhost port=5432 dbname=db sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDSN, gotWarning := ensureSSLMode(tt.dsn)
			if gotDSN != tt.wantDSN {
				t.Errorf("dsn = %q, want %q", gotDSN, tt.wantDSN)
			}
			if (gotWarning != "") != tt.wantWarning {
				t.Errorf("warning = %q, wantWarning = %v", gotWarning, tt.wantWarning)
			}
		})
	}
}

package data

import (
	"testing"
)

func TestOpen(t *testing.T) {

	cfg := DefaultPostgresConfig()
	db, err := Open(cfg)
	
	if err != nil {
		t.Errorf("Expected no error, but got %s", err)
	}
	if db == nil {
		t.Errorf("Expected a DB connection, but got nil")
	}

}

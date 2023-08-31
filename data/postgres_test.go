package data

import "testing"

func TestDefaultPostgresConfig(t *testing.T) {
	cfg := DefaultPostgresConfig()
	if cfg.Host != "localhost" {
		t.Errorf("Expected host to be localhost, but got %s", cfg.Host)
	}
	if cfg.Port != "5431" {
		t.Errorf("Expected port to be 5432, but got %s", cfg.Port)
	}
	if cfg.User != "postgres" {
		t.Errorf("Expected user to be postgres, but got %s", cfg.User)
	}
	if cfg.Password != "postgres" {
		t.Errorf("Expected password to be empty, but got %s", cfg.Password)
	}
	if cfg.Database != "usertest" {
		t.Errorf("Expected database to be postgres, but got %s", cfg.Database)
	}

}

func TestTestPostgresConfig(t *testing.T) {
	cfg := TestPostgresConfig()
	if cfg.Host != "localhost" {
		t.Errorf("Expected host to be localhost, but got %s", cfg.Host)
	}
	if cfg.Port != "5433" {
		t.Errorf("Expected port to be 5432, but got %s", cfg.Port)
	}
	if cfg.User != "postgres" {
		t.Errorf("Expected user to be postgres, but got %s", cfg.User)
	}
	if cfg.Password != "postgres" {
		t.Errorf("Expected password to be empty, but got %s", cfg.Password)
	}
	if cfg.Database != "mockusertest" {
		t.Errorf("Expected database to be postgres, but got %s", cfg.Database)
	}
}

func TestPostgressConfig_String(t *testing.T) {
	cfg := DefaultPostgresConfig()
	got := cfg.String()
	want := "host=localhost port=5431 user=postgres password=postgres dbname=usertest sslmode=disable"
	if got != want {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

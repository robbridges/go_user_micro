package data

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgressConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMODE  string
}

func DefaultPostgresConfig() PostgressConfig {
	return PostgressConfig{
		Host:     "localhost",
		Port:     "5431",
		User:     "postgres",
		Password: "postgres",
		Database: "usertest",
		SSLMODE:  "disable",
	}
}

func Open(config PostgressConfig) (*sql.DB, error) {
	db, err := sql.Open(
		"pgx",
		config.String(),
	)

	if err != nil {
		return nil, fmt.Errorf("error Opening DB: %w", err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("Failed to ping the database:", err)
	}

	return db, nil
}

func (cfg PostgressConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMODE)
}

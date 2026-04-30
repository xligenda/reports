package app

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func DBConfigFromEnv() *DBConfig {
	return &DBConfig{
		Host:     MustEnv("DB_HOST"),
		Port:     MustEnvInt("DB_PORT"),
		User:     MustEnv("DB_USER"),
		Password: MustEnv("DB_PASSWORD"),
		DBName:   MustEnv("DB_NAME"),
		SSLMode:  MustEnv("DB_SSLMODE"),
	}
}

func (a *App) InitDB(cfg *DBConfig) error {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	const maxRetries = 20
	var (
		db  *sqlx.DB
		err error
	)

	for i := range maxRetries {
		db, err = sqlx.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to open DB connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		if err = db.Ping(); err == nil {
			log.Printf("Database connection established after %d attempts", i+1)
			break
		}

		log.Printf("Database not ready yet (attempt %d/%d): %v", i+1, maxRetries, err)
		db.Close()
		db = nil
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}
	if db == nil {
		return fmt.Errorf("database connection is nil after retries")
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	a.DB = db
	log.Println("Database connected and initialized successfully")
	return nil
}

func (a *App) CloseDB() error {
	if a.DB == nil {
		return fmt.Errorf("database already closed or never initialized")
	}

	log.Println("Closing database connection...")
	if err := a.DB.Close(); err != nil {
		log.Printf("Error while closing database connection: %v", err)
		return fmt.Errorf("failed to close database: %w", err)
	}

	a.DB = nil
	log.Println("Database connection closed successfully")
	return nil
}

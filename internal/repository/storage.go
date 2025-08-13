package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

// Storage представляет слой доступа к данным PostgreSQL
type Storage struct {
	db  *sql.DB
	log *zap.Logger
}

// NewStorage создает новый экземпляр репозитория
func NewStorage(user string, password string, host string, port string, dbname string, sslmode string, log *zap.Logger) (*Storage, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslmode)
	log = log.With(zap.String("type", "Storage"))

	log.Info("Connecting to PostgreSQL database",
		zap.String("dbname", dbname),
		zap.String("user", user),
		zap.String("sslmode", sslmode))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("Failed to open database connection", zap.Error(err))
		return nil, err
	}

	// Проверка соединения
	log.Info("Testing database connection")
	if err := db.Ping(); err != nil {
		log.Error("Failed to ping database", zap.Error(err))
		return nil, err
	}

	log.Info("Successfully connected to database")

	log.Info("Starting database migrations")

	if err := runMigrations(db); err != nil {
		log.Error("Failed to run migrations", zap.Error(err))
		return nil, err
	}
	log.Info("Successfully migrated database")
	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func runMigrations(db *sql.DB) error {
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = ".../../migrations"
	}

	// Преобразуем относительный путь в абсолютный
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	goose.SetDialect("postgres")
	if err := goose.Up(db, absPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Close закрывает соединение с базой данных
func (s *Storage) Close() error {
	s.log.Info("Closing database connection")
	return s.db.Close()
}

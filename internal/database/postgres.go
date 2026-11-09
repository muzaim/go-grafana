package database

import (
	"database/sql"
	"fmt"
	"time"

	"go-up-monitoring/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
	SQLDB *sql.DB
}

func ConnectPostgres(cfg config.DatabaseConfig) (*DB, error) {
	dsn := cfg.DSN()

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &DB{
		DB:    db,
		SQLDB: sqlDB,
	}, nil
}

func (d *DB) Ping() error {
	if d.SQLDB == nil {
		return fmt.Errorf("sql.DB is nil")
	}
	return d.SQLDB.Ping()
}

func (d *DB) Close() error {
	if d.SQLDB == nil {
		return nil
	}
	return d.SQLDB.Close()
}

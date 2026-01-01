package common

import (
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)


func ConnectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Add prefer_simple_protocol=true to DSN to avoid prepared statement caching issues
	if dsn != "" && !strings.Contains(dsn, "prefer_simple_protocol") {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		dsn += separator + "prefer_simple_protocol=true"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   newLogger,
		PrepareStmt:                              false, // Disable prepared statements for PgBouncer compatibility
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
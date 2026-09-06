package database

import (
	"os"
	"path/filepath"

	"restaurant-table-reservation/internal/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Connect(path string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&domain.Table{}, &domain.Reservation{}); err != nil {
		return nil, err
	}

	return db, nil
}

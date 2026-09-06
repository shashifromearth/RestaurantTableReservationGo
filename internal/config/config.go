package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                  string
	DatabasePath          string
	CancellationLeadHours int
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "data/restaurant.db"
	}

	leadHours := 2
	if v := os.Getenv("CANCELLATION_LEAD_HOURS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			leadHours = parsed
		}
	}

	return Config{
		Port:                  port,
		DatabasePath:          dbPath,
		CancellationLeadHours: leadHours,
	}
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"meridian/internal/models"
)

type MariaDBRepository struct {
	db *sql.DB
}

func NewMariaDBRepository(host, port, user, password, dbName string) *MariaDBRepository {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Warn("Failed to open MariaDB connection, continuing without lookup logging", "error", err)
		return &MariaDBRepository{db: db}
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		slog.Warn("MariaDB unavailable at startup, continuing without lookup logging", "error", err)
	}

	return &MariaDBRepository{db: db}
}

const insertLookupLog = `
	INSERT INTO lookup_logs (ip, requested_at, status, ip_api_duration_ms, country, country_code, city, region_name)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`

func (r *MariaDBRepository) LogLookup(ctx context.Context, ip string, status string, durationMs *int, loc *models.Location) error {
	var country, countryCode, city, regionName *string
	if loc != nil {
		country, countryCode, city, regionName = &loc.Country, &loc.CountryCode, &loc.City, &loc.Region
	}

	_, err := r.db.ExecContext(ctx, insertLookupLog, ip, time.Now().UTC(), status, durationMs, country, countryCode, city, regionName)
	return err
}

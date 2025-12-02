package store

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultMaxOpenConns = 20

type SQL struct {
	Connection *pgxpool.Pool
}

func NewSQL(dsn string) (*SQL, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("store: empty data source name")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Printf("store: failed to parse config dsn=%s err=%v", sanitizeDSN(dsn), err)
		return nil, fmt.Errorf("store: parse config failed: err=%w", err)
	}

	config.MaxConns = defaultMaxOpenConns
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Printf("store: connection failed dsn=%s err=%v", sanitizeDSN(dsn), err)
		return nil, fmt.Errorf("store: connect failed: err=%w", err)
	}

	if err = db.Ping(context.Background()); err != nil {
		log.Printf("store: ping failed dsn=%s err=%v", sanitizeDSN(dsn), err)
		return nil, fmt.Errorf("store: ping failed: err=%w", err)
	}

	log.Printf("store: connected dsn=%s", sanitizeDSN(dsn))

	return &SQL{Connection: db}, nil
}

func sanitizeDSN(dsn string) string {
	idxScheme := strings.Index(dsn, "://")
	if idxScheme < 0 {
		return dsn
	}
	rest := dsn[idxScheme+3:]
	at := strings.Index(rest, "@")
	if at < 0 {
		return dsn
	}
	return dsn[:idxScheme+3] + "***" + rest[at:]
}

package store

import (
	_ "database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
)

const defaultMaxOpenConns = 20

type SQL struct {
	Connection *sqlx.DB
}

func NewSQL(dsn string) (*SQL, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("store: empty data source name")
	}

	if !strings.Contains(dsn, "://") {
		return nil, fmt.Errorf("store: invalid data source name: %s", dsn)
	}

	driver := strings.ToLower(strings.SplitN(dsn, "://", 2)[0])
	if driver == "" {
		return nil, fmt.Errorf("store: unable to detect driver from dsn: %s", sanitizeDSN(dsn))
	}

	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		log.Printf("store: connection failed driver=%s dsn=%s err=%v", driver, sanitizeDSN(dsn), err)
		return nil, fmt.Errorf("store: connect failed: driver=%s err=%w", driver, err)
	}

	db.SetMaxOpenConns(defaultMaxOpenConns)

	log.Printf("store: connected driver=%s dsn=%s", driver, sanitizeDSN(dsn))

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

package store

import (
	_ "database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
)

// Supported database connection strings:
// mysql://username:password@localhost:3306/dbname?tls=true
// sqlite3://username:password@localhost:3306/dbname?tls=true
// postgres://username:password@localhost:5432/dbname?sslmode=disable&search_path=public
// oracle://username:password@:0/?connstr=(description=(address=(protocol=tcp)(host=localhost)(port=1521))(connect_data=(server=dedicated)(sid=dbname)))&persist security info=true&ssl=enable&ssl verify=false
// etc.

const defaultMaxOpenConns = 20

// SQL wraps a sqlx.DB connection pool.
type SQL struct {
	Connection *sqlx.DB
}

// NewSQL creates and returns a new SQL store connected to the provided DSN.
// It validates the DSN format, does basic driver recognition, sets sensible
// connection pool defaults, and logs structured (lowercase key=value) messages.
// Returns a pointer to SQL on success or an error on failure.
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

	// Connect using sqlx which will open and verify the connection.
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		log.Printf("store: connection failed driver=%s dsn=%s err=%v", driver, sanitizeDSN(dsn), err)
		return nil, fmt.Errorf("store: connect failed: driver=%s err=%w", driver, err)
	}

	// Configure connection pool limits.
	db.SetMaxOpenConns(defaultMaxOpenConns)

	log.Printf("store: connected driver=%s dsn=%s", driver, sanitizeDSN(dsn))

	return &SQL{Connection: db}, nil
}

// sanitizeDSN masks credential information in a dsn to avoid logging secrets.
// It replaces the section between "://" and "@" with "***" when both are present.
func sanitizeDSN(dsn string) string {
	idxScheme := strings.Index(dsn, "://")
	if idxScheme < 0 {
		return dsn
	}
	rest := dsn[idxScheme+3:]
	at := strings.Index(rest, "@")
	if at < 0 {
		// No credentials part detected
		return dsn
	}
	// Build masked DSN: keep scheme, replace credentials with "***", keep host/params
	return dsn[:idxScheme+3] + "***" + rest[at:]
}

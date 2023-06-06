package store

import (
	"strings"

	_ "database/sql"
	"github.com/golang-migrate/migrate/v4"
	// _ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	//_ "github.com/sijms/go-ora/v2"
)

// postgres://username:password@localhost:5432/dbname?sslmode=disable
// mongodb://username:password@localhost:27017/?retryWrites=true&w=majority&tls=false
// oracle://username:password@:0/?connstr=(description=(address=(protocol=tcp)(host=localhost)(port=1521))(connect_data=(server=dedicated)(sid=dbname)))&persist security info=true&ssl=enable&ssl verify=false

type Database struct {
	dataSourceName string
	Client         *sqlx.DB
}

func NewDatabase(url string) (database *Database, err error) {
	database = &Database{
		dataSourceName: url,
	}
	database.Client, err = database.connection()

	return
}

// connection established connection to a database instance using provided URI and auth credentials.
func (s Database) connection() (client *sqlx.DB, err error) {
	if !strings.Contains(s.dataSourceName, "://") {
		err = errors.New("sql: undefined data source name " + s.dataSourceName)
		return
	}
	driverName := strings.ToLower(strings.Split(s.dataSourceName, "://")[0])

	client, err = sqlx.Connect(driverName, s.dataSourceName)
	if err != nil {
		return
	}
	client.SetMaxOpenConns(20)

	return
}

func (s Database) Migrate() (err error) {
	migrations, err := migrate.New("file://migrations", s.dataSourceName)
	if err != nil {
		return
	}

	if err = migrations.Up(); err != nil && err != migrate.ErrNoChange {
		return
	}

	return
}

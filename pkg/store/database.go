package store

import (
	_ "database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	// _ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	//_ "github.com/sijms/go-ora/v2"
)

// postgres://username:password@localhost:5432/dbname?sslmode=disable
// mongodb://username:password@localhost:27017/?retryWrites=true&w=majority&tls=false
// oracle://username:password@:0/?connstr=(description=(address=(protocol=tcp)(host=localhost)(port=1521))(connect_data=(server=dedicated)(sid=dbname)))&persist security info=true&ssl=enable&ssl verify=false

type Database struct {
	schema         string
	driverName     string
	dataSourceName string

	Client *sqlx.DB
}

// NewDatabase established connection to a database instance using provided URI and auth credentials.
func NewDatabase(schema, dataSourceName string) (database *Database, err error) {
	database = &Database{
		schema:         schema,
		dataSourceName: dataSourceName,
	}
	database.Client, err = database.connection()
	if err != nil {
		return
	}

	err = database.createSchema()
	if err != nil {
		return
	}

	return
}

func (s *Database) Migrate() (err error) {
	migrations, err := migrate.New("file://migrations", s.dataSourceName)
	if err != nil {
		return
	}

	if err = migrations.Up(); err != nil && err != migrate.ErrNoChange {
		return
	}

	return nil
}

func (s *Database) connection() (client *sqlx.DB, err error) {
	err = s.parseDSN()
	if err != nil {
		return
	}

	client, err = sqlx.Connect(s.driverName, s.dataSourceName)
	if err != nil {
		return
	}
	client.SetMaxOpenConns(20)

	return
}

func (s *Database) parseDSN() (err error) {
	if !strings.Contains(s.dataSourceName, "://") {
		err = errors.New("sql: undefined data source name " + s.dataSourceName)
		return
	}
	s.driverName = strings.ToLower(strings.Split(s.dataSourceName, "://")[0])

	source, err := url.Parse(s.dataSourceName)
	if err != nil {
		return
	}
	sourceQuery := source.Query()

	if s.schema != "" {
		sourceQuery.Set("search_path", s.schema)
		source.RawQuery = sourceQuery.Encode()
		s.dataSourceName = source.String()
	}

	return
}

func (s *Database) createSchema() (err error) {
	if s.schema == "" {
		return
	}

	switch s.driverName {
	case "postgres":
		query := make([]string, 0)
		query = append(query, "BEGIN")
		query = append(query, "SET TIMEZONE='Asia/Almaty'")
		query = append(query, "SET TIME ZONE 'Asia/Almaty'")
		query = append(query, "SET TIMEZONE TO 'Asia/Almaty'")
		query = append(query, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", s.schema))
		query = append(query, "COMMIT")

		_, err = s.Client.Exec(strings.Join(query, ";"))
		if err != nil {
			return
		}
	}
	return
}

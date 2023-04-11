package database

import (
	"strings"

	_ "database/sql"
	//_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	//_ "github.com/sijms/go-ora/v2"
)

// postgres://username:password@localhost:5432/dbname?sslmode=disable
// mongodb://username:password@localhost:27017/?retryWrites=true&w=majority&tls=false
// oracle://username:password@:0/?connstr=(description=(address=(protocol=tcp)(host=localhost)(port=1521))(connect_data=(server=dedicated)(sid=dbname)))&persist security info=true&ssl=enable&ssl verify=false

// New established connection to a database instance using provided URI and auth credentials.
func New(dataSourceName string) (db *sqlx.DB, err error) {
	if !strings.Contains(dataSourceName, "://") {
		err = errors.New("sql: undefined data source name " + dataSourceName)
		return
	}
	driverName := strings.ToLower(strings.Split(dataSourceName, "://")[0])

	db, err = sqlx.Connect(driverName, dataSourceName)
	if err != nil {
		return
	}
	db.SetMaxOpenConns(20)

	switch driverName {
	case "postgres":
		queries := make([]string, 0)
		queries = append(queries, "BEGIN")
		queries = append(queries, "SET TIMEZONE='Asia/Almaty'")
		queries = append(queries, "SET TIME ZONE 'Asia/Almaty'")
		queries = append(queries, "SET TIMEZONE TO 'Asia/Almaty'")
		queries = append(queries, "COMMIT")

		_, err = db.Exec(strings.Join(queries, ";"))
		if err != nil {
			return
		}
	}

	return
}

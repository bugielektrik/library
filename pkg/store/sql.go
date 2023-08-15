package store

import (
	_ "database/sql"
	"errors"
	"strings"

	//_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	//_ "github.com/sijms/go-ora/v2"
)

// postgres://username:password@localhost:5432/dbname?sslmode=disable&search_path=public
// mongodb://username:password@localhost:27017/?retryWrites=true&w=majority&tls=false
// oracle://username:password@:0/?connstr=(description=(address=(protocol=tcp)(host=localhost)(port=1521))(connect_data=(server=dedicated)(sid=dbname)))&persist security info=true&ssl=enable&ssl verify=false

type SQL struct {
	Connection *sqlx.DB
}

func NewSQL(dataSourceName string) (store SQL, err error) {
	if !strings.Contains(dataSourceName, "://") {
		err = errors.New("store: undefined data source name " + dataSourceName)
		return
	}
	driverName := strings.ToLower(strings.Split(dataSourceName, "://")[0])

	store.Connection, err = sqlx.Connect(driverName, dataSourceName)
	if err != nil {
		return
	}
	store.Connection.SetMaxOpenConns(20)

	return
}

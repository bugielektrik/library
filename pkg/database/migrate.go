package database

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(dataSourceName string) error {
	// preparing migration
	m, err := migrate.New("file://migrations", dataSourceName)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

package store

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return fmt.Errorf("store: empty data source name")
	}

	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return fmt.Errorf("store: invalid data source name: %w", err)
	}

	driver := strings.ToLower(strings.Split(u.Scheme, "+")[0])
	migrationsPath := fmt.Sprintf("file://migrations/%s", driver)

	log.Printf("migrate: start driver=%s host=%s path=%s", driver, u.Host, migrationsPath)

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("migrate: new: %w", err)
	}

	defer func() {
		serr, derr := m.Close()
		if derr != nil || serr != nil {
			log.Printf("migrate: close error: serr=%v, derr=%v", serr, derr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("migrate: no-change driver=%s", driver)
			return nil
		}
		return fmt.Errorf("migrate: up: %w", err)
	}

	log.Printf("migrate: applied driver=%s", driver)
	return nil
}

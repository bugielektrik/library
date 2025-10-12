package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"library-service/internal/infrastructure/store"
)

func main() {
	var (
		direction string
		steps     int
	)

	flag.StringVar(&direction, "direction", "up", "Migration direction: up, down, or version")
	flag.IntVar(&steps, "steps", 0, "Number of migration steps (0 = all)")
	flag.Parse()

	// Get store DSN from environment
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN environment variable is required")
	}

	fmt.Printf("Running migrations: direction=%s, steps=%d\n", direction, steps)

	switch direction {
	case "up":
		if err := store.RunMigrations(dsn); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("✓ Migrations completed successfully")

	case "down":
		fmt.Println("Migration down not yet implemented")
		fmt.Println("To rollback, manually run migration down files")
		os.Exit(1)

	case "version":
		fmt.Println("Migration version check not yet implemented")
		os.Exit(1)

	default:
		log.Fatalf("Unknown migration direction: %s", direction)
	}
}

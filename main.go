package main

import (
	"context"
	"flag"
	"fmt"
	"go-migrations/migrator"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .end file")
	}

	var opts migrator.Options
	flag.BoolVar(&opts.All, "all", false, "Apply all pending migrations")
	flag.BoolVar(&opts.Reset, "reset", false, "Revert all migrations")
	flag.IntVar(&opts.To, "to", 0, "Migrate to a specific version")
	flag.IntVar(&opts.Up, "up", 0, "Migrate up a certain number of versions")
	flag.IntVar(&opts.Down, "down", 0, "Migrate down a certain number of versions")
	flag.Parse()

	ctx := context.Background()
	lastestMigration, err := migrator.Migrate(ctx, &opts)
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	fmt.Printf("DB currently at v%v: %q \n", lastestMigration.Version, lastestMigration.Name)
}

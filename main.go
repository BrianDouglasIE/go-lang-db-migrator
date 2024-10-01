package main

import (
	"flag"
	"go-migrations/pkg/migrations"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .end file")
	}

	to := flag.Int("to", 0, "Migrate to a specific version")
	up := flag.Int("up", 0, "Migrate up a certain number of versions")
	down := flag.Int("down", 0, "Migrate down a certain number of versions")
	clean := flag.Bool("clean", false, "Revert all migrations")
	all := flag.Bool("all", false, "Apply all pending migrations")

	flag.Parse()

	opts := migrations.MigrationOptions{All: *all, To: *to, Up: *up, Down: *down, Clean: *clean}
	err = migrations.Migrate(&opts)
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
}

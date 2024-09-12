package main

import (
	"errors"
	"flag"
	"fmt"
	"simple-chat/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var action string

	flag.StringVar(&action, "action", "", "migrations action")
	flag.Parse()

	if action != "up" && action != "down" {
		fmt.Println("action flag is required (example: --action=up)")
		return
	}

	cfg := config.MustLoad()

	fmt.Println(action)

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s&x-migrations-table=migrations",
		cfg.Database.User,
		cfg.Database.Pass,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	fmt.Println("path: ", cfg.MigrationsPath)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.MigrationsPath),
		dsn,
	)

	if err != nil {
		panic(err)
	}

	if action == "up" {
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}
			panic(err)
		}
	}

	if action == "down" {
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}
			panic(err)
		}
	}

	fmt.Println("migrations applied")
}

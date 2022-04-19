package main

import "github.com/golang-migrate/migrate/v4"

func runMigrations(connectionInfo string) error {
	m, err := migrate.New(
		"file://db/migrations",
		connectionInfo)
	if err != nil {
		return err
	}
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

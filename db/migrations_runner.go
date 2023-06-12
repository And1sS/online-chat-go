package db

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"io/fs"
)

func RunMigrations(fs fs.FS, path string, dbUrl string) error {
	driver, err := iofs.New(fs, path)
	if err != nil {
		return err
	}

	runner, err := migrate.NewWithSourceInstance("iofs", driver, dbUrl)
	if err != nil {
		return err
	}

	err = runner.Up()
	if err != nil {
		return err
	}

	return nil
}

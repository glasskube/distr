package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"strings"

	"github.com/glasskube/cloud/internal/env"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

//go:embed *
var fs embed.FS

type Logger struct {
	*zap.Logger
}

// Printf implements migrate.Logger.
func (l *Logger) Printf(format string, v ...interface{}) {
	if strings.HasPrefix(format, "error") {
		l.Sugar().Errorf(format, v...)
	} else {
		l.Sugar().Infof(format, v...)
	}
}

// Verbose implements migrate.Logger.
func (l *Logger) Verbose() bool {
	return l.Level() == zap.DebugLevel
}

var _ migrate.Logger = &Logger{}

func Up(log *zap.Logger) (err error) {
	db, err := sql.Open("pgx", env.DatabaseUrl())
	if err != nil {
		return err
	}
	defer func() { multierr.AppendInto(&err, db.Close()) }()
	if instance, err := getInstance(db, log); err != nil {
		return err
	} else if err := instance.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		log.Info("migrations completed", zap.Error(err))
	}
	return nil
}

func Down(log *zap.Logger) (err error) {
	db, err := sql.Open("pgx", env.DatabaseUrl())
	if err != nil {
		return err
	}
	defer func() { multierr.AppendInto(&err, db.Close()) }()
	if instance, err := getInstance(db, log); err != nil {
		return err
	} else if err := instance.Down(); err != nil {
		return err
	}
	return nil
}

func getInstance(db *sql.DB, log *zap.Logger) (*migrate.Migrate, error) {
	if driver, err := postgres.WithInstance(db, &postgres.Config{}); err != nil {
		return nil, err
	} else if sourceInstance, err := iofs.New(fs, "."); err != nil {
		return nil, err
	} else if instance, err := migrate.NewWithInstance("", sourceInstance, "cloud", driver); err != nil {
		return nil, err
	} else {
		instance.Log = &Logger{log}
		return instance, nil
	}
}

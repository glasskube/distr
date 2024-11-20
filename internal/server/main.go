package server

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	dbPool *pgxpool.Pool
}

var instance *server

func Init() error {
	if instance != nil {
		return errors.New("server already initialized")
	}

	instance = &server{}

	var err error
	// TODO read DB connection options from environment here or get it passed as param
	instance.dbPool, err = pgxpool.New(context.Background(), "postgres://local:local@localhost:5432/glasskube")
	if err != nil {
		return err
	}
	return nil
}

func Shutdown() error {
	if instance == nil {
		return errors.New("server not yet initialized")
	}
	instance.dbPool.Close()
	return nil
}

func getDbPool() *pgxpool.Pool {
	if instance == nil {
		panic("server not initialized")
	}
	if instance.dbPool == nil {
		panic("db not initialized")
	}
	return instance.dbPool
}

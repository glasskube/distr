package server

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type server struct {
	dbPool *pgxpool.Pool
	logger *zap.Logger
}

var instance *server

func Init() error {
	if instance != nil {
		return errors.New("server already initialized")
	}

	instance = &server{}
	instance.logger = createLogger()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(instance.logger)

	instance.logger.Info("initializing server")

	// TODO read DB connection options from environment here or get it passed as param
	dbConfig, err := pgxpool.ParseConfig("postgres://local:local@localhost:5432/glasskube")
	if err != nil {
		return err
	}
	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		typeNames := []string{"DEPLOYMENT_TYPE"}
		if pgTypes, err := conn.LoadTypes(ctx, typeNames); err != nil {
			return err
		} else {
			conn.TypeMap().RegisterTypes(pgTypes)
			return nil
		}
	}
	instance.dbPool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		instance.logger.Error("cannot set up db pool", zap.Error(err))
		return err
	} else if conn, err := instance.dbPool.Acquire(context.Background()); err != nil {
		// this actually checks whether the DB can be connected to
		instance.logger.Error("cannot acquire connection", zap.Error(err))
		return err
	} else {
		conn.Release()
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

func createLogger() *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	return zap.Must(config.Build())
}

func getLogger() *zap.Logger {
	if instance == nil {
		panic("server not initialized")
	}
	if instance.logger == nil {
		panic("logger not initialized")
	}
	return instance.logger
}

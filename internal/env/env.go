package env

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	envDevelopment = "development"
)

var (
	currentEnv  string
	databaseUrl string
	jwtSecret   []byte
	host        string
)

func init() {
	currentEnv = os.Getenv("GLASSKUBE_ENV")
	if currentEnv == "" {
		currentEnv = envDevelopment
	}
	fmt.Fprintf(os.Stderr, "environment=%v\n", currentEnv)
	if err := godotenv.Load(".env." + currentEnv + ".local"); err != nil && IsDev() {
		fmt.Fprintf(os.Stderr, "environment not loaded: %v\n", err)
	}
	databaseUrl = os.Getenv("DATABASE_URL")
	if decoded, err := base64.StdEncoding.DecodeString(os.Getenv("JWT_SECRET")); err != nil {
		panic(err)
	} else {
		jwtSecret = decoded
	}
	host = os.Getenv("GLASSKUBE_HOST")
	if host == "" {
		panic(errors.New("can't start, GLASSKUBE_HOST not set"))
	}
}

func DatabaseUrl() string {
	return databaseUrl
}

func JWTSecret() []byte {
	return jwtSecret
}

func Host() string { return host }

func IsDev() bool {
	return currentEnv == envDevelopment
}

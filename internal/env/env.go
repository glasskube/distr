package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	envDevelopment = "development"
)

var (
	currentEnv  = ""
	databaseUrl = ""
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
}

func DatabaseUrl() string {
	return databaseUrl
}

func IsDev() bool {
	return currentEnv == envDevelopment
}

package main

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*

agent docker compose file (response of /api/connect?token=...) will look like this:

name: glasskube-agent
services:
	agent:
		image: 'glasskube-agent'
		environment:
			GK_AGENT_TOKEN: JWT
			GK_ENDPOINT: https://glasskube.cloud/api/deployment-targets/hardcoded-id/latest-deployment/download
*/

/*
* wie bei password mit salt und hash
* access key id ist die deployment target id
* on the fly generieren beim instructions anzeigen (nur hier ist das generierte "cleartext" passwort verfügbar zum anzeigen)
* nur an dem punkt gibts das decrypted secret das man im browser anzeigen kann
* zusätzlichen key generieren wird noch nicht supported – wenn das deployment target schon einen status hat, nicht überschreiben derweil

* beim deployment target in der DB steht der encrypted key
 */

func main() {
	token := getFromEnvOrDie("GK_AGENT_TOKEN")
	endpoint := getFromEnvOrDie("GK_ENDPOINT") // TODO /deployment-targets/.../latest-deployment/download

	logger := createLogger()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
		<-sigint
		fmt.Println("ok bye")
		os.Exit(0)
	}()

	client := http.Client{}

	for {
		if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
			logger.Error("failed to create request", zap.Error(err))
		} else {
			req.Header.Set("Authorization", "Bearer "+token)

			if resp, err := client.Do(req); err != nil {
				logger.Error("failed to execute request", zap.Error(err))
			} else {
				if body, err := io.ReadAll(resp.Body); err != nil {
					logger.Error("failed to read response body", zap.Error(err))
				} else {
					fmt.Fprintf(os.Stderr, "%v\n", string(body))
					// TODO apply
					// TODO report status
					if resp.StatusCode != http.StatusOK {
						logger.Error("status code not OK", zap.Int("code", resp.StatusCode), zap.ByteString("body", body))
					}
				}
				_ = resp.Body.Close()
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func getFromEnvOrDie(arg string) string {
	value := os.Getenv(arg)
	if value == "" {
		fmt.Fprintf(os.Stderr, "Cannot start glasskube agent: %v is missing.", arg)
		os.Exit(1)
	}
	return value
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

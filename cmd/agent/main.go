package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	targetId := getFromEnvOrDie("GK_TARGET_ID")
	targetSecret := getFromEnvOrDie("GK_TARGET_SECRET")
	resourceEndpoint := getFromEnvOrDie("GK_RESOURCE_ENDPOINT")
	statusEndpoint := getFromEnvOrDie("GK_STATUS_ENDPOINT")

	logger := createLogger()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
		<-sigint
		cancel()
	}()

	client := http.Client{}

	for ctx.Err() == nil {
		if req, err := http.NewRequest(http.MethodGet, resourceEndpoint, nil); err != nil {
			logger.Error("failed to create request", zap.Error(err))
		} else {
			req.SetBasicAuth(targetId, targetSecret)
			if resp, err := client.Do(req); err != nil {
				logger.Error("failed to execute request", zap.Error(err))
			} else if resp.StatusCode != http.StatusOK {
				logger.Warn("status code not OK, will not apply", zap.Int("code", resp.StatusCode))
			} else {
				status := make(map[string]string)
				cmd := exec.Command("docker", "compose", "-f", "-", "up", "-d", "--quiet-pull")
				cmd.Stdin = resp.Body
				out, cmdErr := cmd.CombinedOutput()
				if cmdErr != nil {
					status["error"] = cmdErr.Error()
				}
				status["output"] = string(out)
				logger.Debug("docker compose returned", zap.String("output", string(out)), zap.Error(cmdErr))

				correlationID := resp.Header.Get("X-Resource-Correlation-ID")
				if statusJson, err := json.Marshal(status); err != nil {
					logger.Error("failed to marshal status JSON", zap.Error(err))
				} else if statusReq, err :=
					http.NewRequest(http.MethodPost, statusEndpoint, bytes.NewReader(statusJson)); err != nil {
					logger.Error("failed to create status request", zap.Error(err))
				} else if correlationID != "" {
					statusReq.Header.Set("Content-Type", "application/json")
					statusReq.Header.Set("X-Resource-Correlation-ID", correlationID)
					statusReq.SetBasicAuth(targetId, targetSecret)
					if statusResp, err := client.Do(statusReq); err != nil {
						logger.Error("failed to execute status request", zap.Error(err))
					} else if statusResp.StatusCode != http.StatusOK {
						logger.Info("response code of status request was not OK", zap.Int("code", statusResp.StatusCode))
					}
				}
			}
		}

		sleepDone := make(chan struct{}, 1)
		go func() {
			time.Sleep(5 * time.Second)
			sleepDone <- struct{}{}
		}()
		select {
		case <-sleepDone:
		case <-ctx.Done():
		}
	}
}

func getFromEnvOrDie(arg string) string {
	value, ok := os.LookupEnv(arg)
	if !ok {
		fmt.Fprintf(os.Stderr, "Cannot start glasskube agent: %v is missing.", arg)
		os.Exit(1)
	}
	return value
}

func createLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zap.Must(config.Build())
}

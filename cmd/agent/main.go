package main

import (
	"context"
	"fmt"
	"io"
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
	accessKeyId := getFromEnvOrDie("GK_ACCESS_KEY_ID")
	accessKeySecret := getFromEnvOrDie("GK_ACCESS_KEY_SECRET")
	endpoint := getFromEnvOrDie("GK_ENDPOINT")

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
		if req, err := http.NewRequest("GET", endpoint, nil); err != nil {
			logger.Error("failed to create request", zap.Error(err))
		} else {
			req.SetBasicAuth(accessKeyId, accessKeySecret)

			if resp, err := client.Do(req); err != nil {
				logger.Error("failed to execute request", zap.Error(err))
			} else {
				if body, err := io.ReadAll(resp.Body); err != nil {
					logger.Error("failed to read response body", zap.Error(err))
				} else {
					fmt.Fprintf(os.Stderr, "---response body: \n%v\n---\n", string(body))

					err = os.WriteFile("/tmp/compose.yaml", body, 0644)
					if err != nil {
						logger.Error("failed to write temp file", zap.Error(err))
					} else if out, err := exec.Command(
						"docker", "compose", "-f", "/tmp/compose.yaml", "up", "-d", "--quiet-pull").CombinedOutput(); err != nil {
						logger.Error("failed", zap.Error(err), zap.String("out", string(out)))
					} else {
						fmt.Fprintf(os.Stderr, "---compose up output: \n%v\n---\n", string(out))
					}
					// TODO report deployment status
					if resp.StatusCode != http.StatusOK {
						logger.Error("status code not OK", zap.Int("code", resp.StatusCode), zap.ByteString("body", body))
					}
				}
				_ = resp.Body.Close()
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

package main

import (
	"encoding/base64"
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
			auth := fmt.Sprintf("%s:%s", accessKeyId, accessKeySecret)
			encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Set("Authorization", "Basic "+encodedAuth)

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
						"docker", "compose", "-f", "/tmp/compose.yaml", "up", "-d").CombinedOutput(); err != nil {
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

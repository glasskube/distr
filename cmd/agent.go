package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	token := getOrQuit("GK_AGENT_TOKEN")
	endpoint := getOrQuit("GK_ENDPOINT") // TODO /deployment-targets/.../latest-deployment/download

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
			fmt.Fprintf(os.Stderr, "failed to create request: %v\n", err)
		} else {
			req.Header.Set("Authorization", "Bearer "+token)

			resp, reqErr := client.Do(req)
			if reqErr != nil {
				fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
			} else {
				if resp.StatusCode != http.StatusOK {
					fmt.Fprintf(os.Stderr, "status code not OK: %v\n", resp.StatusCode)
				}
				if body, readErr := io.ReadAll(resp.Body); readErr != nil {
					fmt.Fprintf(os.Stderr, "failed to read response body: %v\n", readErr)
				} else {
					fmt.Fprintf(os.Stderr, "%v\n", string(body))
				}
				_ = resp.Body.Close()
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func getOrQuit(arg string) string {
	value := os.Getenv(arg)
	if value == "" {
		fmt.Fprintf(os.Stderr, "Cannot start glasskube agent: %v is missing.", arg)
		os.Exit(1)
	}
	return value
}

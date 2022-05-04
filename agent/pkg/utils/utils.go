package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
)

var (
	StartTime int64 // global
)

// StartServer starts the server with a graceful shutdown
func StartServer(app *gin.Engine) {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals,
		os.Interrupt,    // this catch ctrl + c
		syscall.SIGTSTP, // this catch ctrl + z
	)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: app,
	}

	go func() {
		<-signals
		logger.Log.Infof("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			logger.Log.Errorf("%v", err)
		}
		os.Exit(0)
	}()

	// Run server.
	logger.Log.Infof("Starting the server...")
	if err := app.Run(fmt.Sprintf(":%d", shared.DefaultApiServerPort)); err != nil {
		logger.Log.Errorf("Server is not running! Reason: %v", err)
	}
}

func CheckErr(e error) {
	if e != nil {
		logger.Log.Errorf("%v", e)
	}
}

func SetHostname(address, newHostname string) string {
	replacedUrl, err := url.Parse(address)
	if err != nil {
		logger.Log.Errorf("error replacing hostname to %s in address %s, returning original %v", newHostname, address, err)
		return address
	}
	replacedUrl.Host = newHostname
	return replacedUrl.String()

}

func ReadJsonFile(filePath string, value interface{}) error {
	if content, err := ioutil.ReadFile(filePath); err != nil {
		return err
	} else {
		if err = json.Unmarshal(content, value); err != nil {
			return err
		}
	}

	return nil
}

func SaveJsonFile(filePath string, value interface{}) error {
	if data, err := json.Marshal(value); err != nil {
		return err
	} else {
		if err = ioutil.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func UniqueStringSlice(s []string) []string {
	uniqueSlice := make([]string, 0)
	uniqueMap := map[string]bool{}

	for _, val := range s {
		if uniqueMap[val] {
			continue
		}
		uniqueMap[val] = true
		uniqueSlice = append(uniqueSlice, val)
	}

	return uniqueSlice
}

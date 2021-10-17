package utils

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared/logger"
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
		_ = <-signals
		logger.Log.Infof("Shutting down...")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		_ = srv.Shutdown(ctx)
		os.Exit(0)
	}()

	// Run server.
	logger.Log.Infof("Starting the server...")
	if err := app.Run(":8899"); err != nil {
		logger.Log.Errorf("Server is not running! Reason: %v", err)
	}
}

func ReverseSlice(data interface{}) {
	value := reflect.ValueOf(data)
	valueLen := value.Len()
	for i := 0; i <= int((valueLen-1)/2); i++ {
		reverseIndex := valueLen - 1 - i
		tmp := value.Index(reverseIndex).Interface()
		value.Index(reverseIndex).Set(value.Index(i))
		value.Index(i).Set(reflect.ValueOf(tmp))
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

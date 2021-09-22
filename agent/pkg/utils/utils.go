package utils

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/romana/rlog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
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
		rlog.Infof("Shutting down...")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		_ = srv.Shutdown(ctx)
		os.Exit(0)
	}()

	// Run server.
	rlog.Infof("Starting the server...")
	if err := app.Run(":8899"); err != nil {
		rlog.Errorf("Server is not running! Reason: %v", err)
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
		rlog.Errorf("%v", e)
	}
}

func SetHostname(address, newHostname string) string {
	replacedUrl, err := url.Parse(address)
	if err != nil {
		rlog.Errorf("error replacing hostname to %s in address %s, returning original %v", newHostname, address, err)
		return address
	}
	replacedUrl.Host = newHostname
	return replacedUrl.String()

}

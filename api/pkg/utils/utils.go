package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

// StartServer starts the server with a graceful shutdown
func StartServer(app *fiber.App) {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals,
		os.Interrupt,  	  // this catch ctrl + c
		syscall.SIGTSTP,  // this catch ctrl + z
	)

	go func() {
		_ = <-signals
		fmt.Println("Shutting down...")
		_ = app.Shutdown()
	}()

	// Run server.
	if err := app.Listen(":8899"); err != nil {
		log.Printf("Oops... Server is not running! Reason: %v", err)
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
		log.Printf("%v", e)
		//panic(e)
	}
}

func SetHostname(address, newHostname string) string {
	replacedUrl, err := url.Parse(address)
	if err != nil{
		log.Printf("error replacing hostname to %s in address %s, returning original %v",newHostname, address, err)
		return address
	}
	replacedUrl.Host = newHostname
	return replacedUrl.String()

}

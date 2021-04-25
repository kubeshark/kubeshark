package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"os/signal"
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


func CheckErr(e error) {
	if e != nil {
		panic(e)
	}
}
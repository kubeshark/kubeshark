package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
)

/// starts the server with a graceful shutdown
func StartServer(app *fiber.App) {
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs,
		os.Interrupt,  	  // this catch ctrl + c
		syscall.SIGTSTP,  // this catch ctrl + z
	)

	go func() {
		_ = <-sigs
		fmt.Println("Shutting down...")
		_ = app.Shutdown()
	}()

	// Run server.
	if err := app.Listen(":8899"); err != nil {
		log.Printf("Oops... Server is not running! Reason: %v", err)
	}
}

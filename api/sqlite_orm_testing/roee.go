package main

import (
	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model  // this will add "id", "created updated inserted times"

	Name string
	Age  int
	Birthday time.Time
}

func main () {
	// github.com/mattn/go-sqlite3
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil{
		panic("ff")
	}

	// Run migrations
	migErr := db.AutoMigrate(&User{})
	if migErr != nil {
		panic("Cannot run migration")
	}

	user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}
	db.Create(&user) // pass pointer of data to Create

}



//func main() {
//	app := fiber.New()
//
//	c := make(chan os.Signal, 1)
//	signal.Notify(c, os.Interrupt)
//	go func() {
//		_ = <-c
//		fmt.Println("Gracefully shutting down...")
//		_ = app.Shutdown()
//	}()
//
//	if err := app.Listen(":3000"); err != nil {
//		log.Panic(err)
//	}
//
//	fmt.Println("Running cleanup tasks...")
//	// Your cleanup tasks go here
//}

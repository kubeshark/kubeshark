package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"mizuserver/src/pkg/models"
)

const (
	DBPath = "./entries.db"
)

var (
	DB = initDataBase(DBPath)
	EntriesCollection = DB.Table("mizu_entries")
)

func initDataBase(databasePath string) *gorm.DB {
	temp, _ := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	_ = temp.AutoMigrate(&models.MizuEntry{}) // this will ensure table is created
	return temp
}
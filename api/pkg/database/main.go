package database

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
)

const (
	DBPath = "./entries.db"
)

var (
	DB = initDataBase(DBPath)
)

const (
	OrderDesc = "desc"
	OrderAsc  = "asc"
	LT        = "lt"
	GT        = "gt"
)

var (
	OperatorToSymbolMapping = map[string]string{
		LT: "<",
		GT: ">",
	}
	OperatorToOrderMapping = map[string]string{
		LT: OrderDesc,
		GT: OrderAsc,
	}
)

func GetEntriesTable() *gorm.DB {
	return DB.Table("mizu_entries")
}

func initDataBase(databasePath string) *gorm.DB {
	go StartEnforcingDatabaseSize( 10 * 1000 * 1000)
	temp, _ := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	_ = temp.AutoMigrate(&models.MizuEntry{}) // this will ensure table is created
	return temp
}


func GetEntriesFromDb(timestampFrom int64, timestampTo int64) []models.MizuEntry {
	order := OrderDesc
	var entries []models.MizuEntry
	GetEntriesTable().
		Where(fmt.Sprintf("timestamp BETWEEN %v AND %v", timestampFrom, timestampTo)).
		Order(fmt.Sprintf("timestamp %s", order)).
		Find(&entries)

	if len(entries) > 0 {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}
	return entries
}


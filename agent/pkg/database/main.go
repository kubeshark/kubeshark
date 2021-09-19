package database

import (
	"fmt"
	"mizuserver/pkg/utils"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	tapApi "github.com/up9inc/mizu/tap/api"
)

const (
	DBPath    = "./entries.db"
	OrderDesc = "desc"
	OrderAsc  = "asc"
	LT        = "lt"
	GT        = "gt"
)

var (
	DB                      *gorm.DB
	IsDBLocked              = false
	OperatorToSymbolMapping = map[string]string{
		LT: "<",
		GT: ">",
	}
	OperatorToOrderMapping = map[string]string{
		LT: OrderDesc,
		GT: OrderAsc,
	}
)

func init() {
	DB = initDataBase(DBPath)
	go StartEnforcingDatabaseSize()
}

func GetEntriesTable() *gorm.DB {
	return DB.Table("mizu_entries")
}

func CreateEntry(entry *tapApi.MizuEntry) {
	if IsDBLocked {
		return
	}
	GetEntriesTable().Create(entry)
}

func initDataBase(databasePath string) *gorm.DB {
	temp, _ := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: &utils.TruncatingLogger{LogLevel: logger.Warn, SlowThreshold: 500 * time.Millisecond},
	})
	_ = temp.AutoMigrate(&tapApi.MizuEntry{}) // this will ensure table is created
	return temp
}

func GetEntriesFromDb(timestampFrom int64, timestampTo int64, protocolName *string) []tapApi.MizuEntry {
	order := OrderDesc
	protocolNameCondition := "1 = 1"
	if protocolName != nil {
		protocolNameCondition = fmt.Sprintf("protocolKey = '%s'", *protocolName)
	}

	var entries []tapApi.MizuEntry
	GetEntriesTable().
		Where(protocolNameCondition).
		Where(fmt.Sprintf("timestamp BETWEEN %v AND %v", timestampFrom, timestampTo)).
		Order(fmt.Sprintf("timestamp %s", order)).
		Find(&entries)

	if len(entries) > 0 {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}
	return entries
}

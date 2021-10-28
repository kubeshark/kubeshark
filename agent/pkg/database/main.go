package database

import (
	"fmt"
	"mizuserver/pkg/utils"
	"time"

	"gorm.io/gorm"

	tapApi "github.com/up9inc/mizu/tap/api"
)

const (
	DBPath     = "./entries.db"
	OrderDesc  = "desc"
	OrderAsc   = "asc"
	LT         = "lt"
	GT         = "gt"
	TimeFormat = "2006-01-02 15:04:05.000000000"
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
}

func GetEntriesTable() *gorm.DB {
	return DB.Table("mizu_entries")
}

// func CreateEntry(entry *tapApi.MizuEntry) {
// 	if IsDBLocked {
// 		return
// 	}
// 	GetEntriesTable().Create(entry)
// }

func GetEntriesFromDb(timeFrom time.Time, timeTo time.Time, protocolName *string) []tapApi.MizuEntry {
	order := OrderDesc
	protocolNameCondition := "1 = 1"
	if protocolName != nil {
		protocolNameCondition = fmt.Sprintf("protocolName = '%s'", *protocolName)
	}

	var entries []tapApi.MizuEntry
	GetEntriesTable().
		Where(protocolNameCondition).
		Where(fmt.Sprintf("created_at BETWEEN '%s' AND '%s'", timeFrom.Format(TimeFormat), timeTo.Format(TimeFormat))).
		Order(fmt.Sprintf("timestamp %s", order)).
		Find(&entries)

	if len(entries) > 0 {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}
	return entries
}

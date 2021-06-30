package controllers

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/martian/har"
	"io/ioutil"
	"log"
	"mizuserver/pkg/database"
	"mizuserver/pkg/models"
	"mizuserver/pkg/utils"
	"mizuserver/pkg/validation"
	"net/http"
	"net/url"
	"time"
)

const (
	OrderDesc = "desc"
	OrderAsc  = "asc"
	LT        = "lt"
	GT        = "gt"
)

var (
	operatorToSymbolMapping = map[string]string{
		LT: "<",
		GT: ">",
	}
	operatorToOrderMapping = map[string]string{
		LT: OrderDesc,
		GT: OrderAsc,
	}
)

func GetEntries(c *fiber.Ctx) error {
	entriesFilter := &models.EntriesFilter{}

	if err := c.QueryParser(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	order := operatorToOrderMapping[entriesFilter.Operator]
	operatorSymbol := operatorToSymbolMapping[entriesFilter.Operator]
	var entries []models.MizuEntry
	database.GetEntriesTable().
		Order(fmt.Sprintf("timestamp %s", order)).
		Where(fmt.Sprintf("timestamp %s %v", operatorSymbol, entriesFilter.Timestamp)).
		Omit("entry"). // remove the "big" entry field
		Limit(entriesFilter.Limit).
		Find(&entries)

	if len(entries) > 0 && order == OrderDesc {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}

	// Convert to base entries
	baseEntries := make([]models.BaseEntryDetails, 0, entriesFilter.Limit)
	for _, entry := range entries {
		baseEntries = append(baseEntries, utils.GetResolvedBaseEntry(entry))
	}

	return c.Status(fiber.StatusOK).JSON(baseEntries)
}

func GetHARs(c *fiber.Ctx) error {
	entriesFilter := &models.HarFetchRequestBody{}
	order := OrderDesc
	if err := c.QueryParser(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	var timestampFrom, timestampTo int64

	if entriesFilter.From < 0 {
		timestampFrom = 0
	} else {
		timestampFrom = entriesFilter.From
	}
	if entriesFilter.To <= 0 {
		timestampTo = time.Now().UnixNano() / int64(time.Millisecond)
	} else {
		timestampTo = entriesFilter.To
	}

	var entries []models.MizuEntry
	database.GetEntriesTable().
		Where(fmt.Sprintf("timestamp BETWEEN %v AND %v", timestampFrom, timestampTo)).
		Order(fmt.Sprintf("timestamp %s", order)).
		Find(&entries)

	if len(entries) > 0 {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}

	harsObject := map[string]*models.ExtendedHAR{}

	for _, entryData := range entries {
		var harEntry har.Entry
		_ = json.Unmarshal([]byte(entryData.Entry), &harEntry)

		sourceOfEntry := entryData.ResolvedSource
		fileName := fmt.Sprintf("%s.har", sourceOfEntry)
		if harOfSource, ok := harsObject[fileName]; ok {
			harOfSource.Log.Entries = append(harOfSource.Log.Entries, &harEntry)
		} else {
			var entriesHar []*har.Entry
			entriesHar = append(entriesHar, &harEntry)
			harsObject[fileName] = &models.ExtendedHAR{
				Log: &models.ExtendedLog{
					Version: "1.2",
					Creator: &models.ExtendedCreator{
						Creator: &har.Creator{
							Name:    "mizu",
							Version: "0.0.2",
						},
						Source: sourceOfEntry,
					},
					Entries: entriesHar,
				},
			}
		}
	}

	retObj := map[string][]byte{}
	for k, v := range harsObject {
		bytesData, _ := json.Marshal(v)
		retObj[k] = bytesData
	}
	buffer := utils.ZipData(retObj)
	return c.Status(fiber.StatusOK).SendStream(buffer)
}

func uploadEntriesImpl(token string, model string, envPrefix string) {
	sleepTime := time.Second * 10

	var timestampFrom int64 = 0

	for {
		timestampTo := time.Now().UnixNano() / int64(time.Millisecond)
		fmt.Printf("Getting entries from %v, to %v\n", timestampFrom, timestampTo)
		entriesArray := getEntriesFromDb(timestampFrom, timestampTo)

		if len(entriesArray) > 0 {
			fmt.Printf("About to upload %v entries\n", len(entriesArray))

			body, jMarshalErr := json.Marshal(entriesArray)
			if jMarshalErr != nil {
				log.Fatal(jMarshalErr)
			}

			var in bytes.Buffer
			w := zlib.NewWriter(&in)
			_, _ = w.Write(body)
			_ = w.Close()
			reqBody := ioutil.NopCloser(bytes.NewReader(in.Bytes()))

			postUrl, _ := url.Parse(fmt.Sprintf("https://traffic.%s/dumpTrafficBulk/%s", envPrefix, model))
			fmt.Println(postUrl)
			req := &http.Request{
				Method: http.MethodPost,
				URL:    postUrl,
				Header: map[string][]string{
					"Content-Encoding": {"deflate"},
					"Content-Type":     {"application/octet-stream"},
					"Guest-Auth":       {token},
				},
				Body: reqBody,
			}
			_, postErr := http.DefaultClient.Do(req)
			if postErr != nil {
				log.Fatal(postErr)
			}
			fmt.Printf("Finish uploading %v entries to %s\n", len(entriesArray), postUrl)

		} else {
			fmt.Println("Nothing to upload")
		}

		fmt.Printf("Sleeping for %v...\n", sleepTime)
		time.Sleep(sleepTime)
		timestampFrom = timestampTo
	}
}

func UploadEntries(c *fiber.Ctx) error {
	entriesFilter := &models.UploadEntriesRequestBody{}
	if err := c.QueryParser(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	if err := validation.Validate(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	go uploadEntriesImpl(entriesFilter.Token, entriesFilter.Model, entriesFilter.Dest)
	return c.Status(fiber.StatusOK).SendString("OK")
}

func GetFullEntries(c *fiber.Ctx) error {
	entriesFilter := &models.HarFetchRequestBody{}
	if err := c.QueryParser(entriesFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	err := validation.Validate(entriesFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	var timestampFrom, timestampTo int64

	if entriesFilter.From < 0 {
		timestampFrom = 0
	} else {
		timestampFrom = entriesFilter.From
	}
	if entriesFilter.To <= 0 {
		timestampTo = time.Now().UnixNano() / int64(time.Millisecond)
	} else {
		timestampTo = entriesFilter.To
	}

	entriesArray := getEntriesFromDb(timestampFrom, timestampTo)
	return c.Status(fiber.StatusOK).JSON(entriesArray)
}

func getEntriesFromDb(timestampFrom int64, timestampTo int64) []har.Entry {
	order := OrderDesc
	var entries []models.MizuEntry
	database.GetEntriesTable().
		Where(fmt.Sprintf("timestamp BETWEEN %v AND %v", timestampFrom, timestampTo)).
		Order(fmt.Sprintf("timestamp %s", order)).
		Find(&entries)

	if len(entries) > 0 {
		// the entries always order from oldest to newest so we should revers
		utils.ReverseSlice(entries)
	}

	entriesArray := make([]har.Entry, 0)
	for _, entryData := range entries {
		var harEntry har.Entry
		_ = json.Unmarshal([]byte(entryData.Entry), &harEntry)
		entriesArray = append(entriesArray, harEntry)
	}
	return entriesArray
}

func GetEntry(c *fiber.Ctx) error {
	var entryData models.EntryData
	database.GetEntriesTable().
		Select("entry", "resolvedDestination").
		Where(map[string]string{"entryId": c.Params("entryId")}).
		First(&entryData)

	var fullEntry har.Entry
	unmarshallErr := json.Unmarshal([]byte(entryData.Entry), &fullEntry)
	utils.CheckErr(unmarshallErr)

	if entryData.ResolvedDestination != "" {
		fullEntry.Request.URL = utils.SetHostname(fullEntry.Request.URL, entryData.ResolvedDestination)
	}

	return c.Status(fiber.StatusOK).JSON(fullEntry)
}

func DeleteAllEntries(c *fiber.Ctx) error {
	database.GetEntriesTable().
		Where("1 = 1").
		Delete(&models.MizuEntry{})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": "Success",
	})

}

func GetGeneralStats(c *fiber.Ctx) error {
	sqlQuery := "SELECT count(*) as count, min(timestamp) as min, max(timestamp) as max from mizu_entries"
	var result struct {
		Count int
		Min   int
		Max   int
	}
	database.GetEntriesTable().Raw(sqlQuery).Scan(&result)
	return c.Status(fiber.StatusOK).JSON(&result)
}

package entries

import (
	"encoding/json"
	"time"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/app"
	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

func GetEntries(entriesRequest *models.EntriesRequest) (*models.EntriesResponse, error) {
	data, meta, err := basenine.Fetch(shared.BasenineHost, shared.BaseninePort,
		entriesRequest.LeftOff, entriesRequest.Direction, entriesRequest.Query,
		entriesRequest.Limit, time.Duration(entriesRequest.TimeoutMs)*time.Millisecond)
	if err != nil {
		return nil, err
	}

	response := &models.EntriesResponse{}
	var dataSlice []interface{}

	for _, row := range data {
		var entry *tapApi.Entry
		err = json.Unmarshal(row, &entry)
		if err != nil {
			return nil, err
		}

		extension := app.ExtensionsMap[entry.Protocol.Name]
		base := extension.Dissector.Summarize(entry)

		dataSlice = append(dataSlice, base)
	}

	var metadata *basenine.Metadata
	err = json.Unmarshal(meta, &metadata)
	if err != nil {
		logger.Log.Debugf("Error recieving metadata: %v", err.Error())
	}

	response.Data = dataSlice
	response.Meta = metadata

	return response, nil
}

func GetEntry(singleEntryRequest *models.SingleEntryRequest, entryId int) (*tapApi.EntryWrapper, error) {
	var entry *tapApi.Entry
	bytes, err := basenine.Single(shared.BasenineHost, shared.BaseninePort, entryId, singleEntryRequest.Query)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &entry)
	if err != nil {
		return nil, err
	}

	extension := app.ExtensionsMap[entry.Protocol.Name]
	base := extension.Dissector.Summarize(entry)
	var representation []byte
	representation, err = extension.Dissector.Represent(entry.Request, entry.Response)
	if err != nil {
		return nil, err
	}

	var rules []map[string]interface{}
	var isRulesEnabled bool
	if entry.Protocol.Name == "http" {
		harEntry, _ := har.NewEntry(entry.Request, entry.Response, entry.StartTime, entry.ElapsedTime)
		_, rulesMatched, _isRulesEnabled := models.RunValidationRulesState(*harEntry, entry.Destination.Name)
		isRulesEnabled = _isRulesEnabled
		inrec, _ := json.Marshal(rulesMatched)
		if err := json.Unmarshal(inrec, &rules); err != nil {
			logger.Log.Error(err)
		}
	}

	return &tapApi.EntryWrapper{
		Protocol:       entry.Protocol,
		Representation: string(representation),
		Data:           entry,
		Base:           base,
		Rules:          rules,
		IsRulesEnabled: isRulesEnabled,
	}, nil
}

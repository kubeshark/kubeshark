package entries

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/app"
	"github.com/up9inc/mizu/agent/pkg/models"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/shared"
	tapApi "github.com/up9inc/mizu/tap/api"
)

type EntriesProvider interface {
	GetEntries(entriesRequest *models.EntriesRequest) ([]*tapApi.EntryWrapper, *basenine.Metadata, error)
	GetEntry(singleEntryRequest *models.SingleEntryRequest, entryId string) (*tapApi.EntryWrapper, error)
}

type BasenineEntriesProvider struct{}

func (e *BasenineEntriesProvider) GetEntries(entriesRequest *models.EntriesRequest) ([]*tapApi.EntryWrapper, *basenine.Metadata, error) {
	data, _, lastMeta, err := basenine.Fetch(shared.BasenineHost, shared.BaseninePort,
		entriesRequest.LeftOff, entriesRequest.Direction, entriesRequest.Query,
		entriesRequest.Limit, time.Duration(entriesRequest.TimeoutMs)*time.Millisecond)
	if err != nil {
		return nil, nil, err
	}

	var dataSlice []*tapApi.EntryWrapper

	for _, row := range data {
		var entry *tapApi.Entry
		err = json.Unmarshal(row, &entry)
		if err != nil {
			return nil, nil, err
		}

		protocol, ok := app.ProtocolsMap[entry.ProtocolId]
		if !ok {
			return nil, nil, fmt.Errorf("protocol not found, protocol: %v", protocol)
		}

		extension, ok := app.ExtensionsMap[protocol.Name]
		if !ok {
			return nil, nil, fmt.Errorf("extension not found, extension: %v", protocol.Name)
		}

		base := extension.Dissector.Summarize(entry)

		dataSlice = append(dataSlice, &tapApi.EntryWrapper{
			Protocol: *protocol,
			Data:     entry,
			Base:     base,
		})
	}

	var metadata *basenine.Metadata
	err = json.Unmarshal(lastMeta, &metadata)
	if err != nil {
		logger.Log.Debugf("Error recieving metadata: %v", err.Error())
	}

	return dataSlice, metadata, nil
}

func (e *BasenineEntriesProvider) GetEntry(singleEntryRequest *models.SingleEntryRequest, entryId string) (*tapApi.EntryWrapper, error) {
	var entry *tapApi.Entry
	bytes, err := basenine.Single(shared.BasenineHost, shared.BaseninePort, entryId, singleEntryRequest.Query)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &entry)
	if err != nil {
		return nil, errors.New(string(bytes))
	}

	protocol, ok := app.ProtocolsMap[entry.ProtocolId]
	if !ok {
		return nil, fmt.Errorf("protocol not found, protocol: %v", protocol)
	}

	extension, ok := app.ExtensionsMap[protocol.Name]
	if !ok {
		return nil, fmt.Errorf("extension not found, extension: %v", protocol.Name)
	}

	base := extension.Dissector.Summarize(entry)
	var representation []byte
	representation, err = extension.Dissector.Represent(entry.Request, entry.Response)
	if err != nil {
		return nil, err
	}

	return &tapApi.EntryWrapper{
		Protocol:       *protocol,
		Representation: string(representation),
		Data:           entry,
		Base:           base,
	}, nil
}

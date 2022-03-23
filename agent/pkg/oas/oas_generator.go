package oas

import (
	"encoding/json"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"
	"net/url"
	"sync"

	"github.com/up9inc/mizu/shared/logger"
)

func RunDBProcessing() {
	serviceSpecs := new(sync.Map)
	c, err := basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
	if err != nil {
		panic(err)
	}

	// Make []byte channels to recieve the data and the meta
	data := make(chan []byte)
	meta := make(chan []byte)

	// Define a function to handle the data stream
	handleDataChannel := func(wg *sync.WaitGroup, c *basenine.Connection, data chan []byte) {
		defer wg.Done()
		for {
			bytes := <-data

			logger.Log.Debugf("Data: %s", bytes)
			e := new(api.Entry)
			err := json.Unmarshal(bytes, e)
			if err != nil {
				continue
			}
			handleEntry(e, serviceSpecs)
		}
	}

	// Define a function to handle the meta stream
	handleMetaChannel := func(c *basenine.Connection, meta chan []byte) {
		for {
			bytes := <-meta

			logger.Log.Debugf("Meta: %s", bytes)
		}
	}

	var wg sync.WaitGroup
	go handleDataChannel(&wg, c, data)
	go handleMetaChannel(c, meta)
	wg.Add(1)

	c.Query("", data, meta)

	wg.Wait()
}

func handleEntry(mizuEntry *api.Entry, specs *sync.Map) {
	entry, err := har.NewEntry(mizuEntry.Request, mizuEntry.Response, mizuEntry.StartTime, mizuEntry.ElapsedTime)
	if err != nil {
		logger.Log.Warningf("Failed to turn MizuEntry %d into HAR Entry: %s", mizuEntry.Id, err)
		return
	}

	dest := mizuEntry.Destination.Name
	if dest == "" {
		dest = mizuEntry.Destination.IP + ":" + mizuEntry.Destination.Port
	}

	u, err := url.Parse(entry.Request.URL)
	if err != nil {
		logger.Log.Errorf("Failed to parse entry URL: %v, err: %v", entry.Request.URL, err)
	}

	val, found := specs.Load(dest)
	var gen *SpecGen
	if !found {
		gen = NewGen(u.Scheme + "://" + dest)
		specs.Store(dest, gen)
	} else {
		gen = val.(*SpecGen)
	}

	entryWSource := EntryWithSource{
		Entry:       *entry,
		Source:      mizuEntry.Source.Name,
		Destination: dest,
		Id:          mizuEntry.Id,
	}

	opId, err := gen.feedEntry(entryWSource)
	if err != nil {
		txt, suberr := json.Marshal(entry)
		if suberr == nil {
			logger.Log.Debugf("Problematic entry: %s", txt)
		}

		logger.Log.Warningf("Failed processing entry %d: %s", mizuEntry.Id, err)
		return
	}

	logger.Log.Debugf("Handled entry %d as opId: %s", mizuEntry.Id, opId) // TODO: set opId back to entry?
}

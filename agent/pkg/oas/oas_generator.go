package oas

import (
	"context"
	"encoding/json"
	"net/url"
	"sync"

	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"

	"github.com/up9inc/mizu/logger"
)

var (
	syncOnce sync.Once
	instance *defaultOasGenerator
)

type OasGenerator interface {
	Start(conn *basenine.Connection)
	Stop()
	IsStarted() bool
	GetServiceSpecs() *sync.Map
	SetEntriesQuery(query string) bool
}

type defaultOasGenerator struct {
	started      bool
	ctx          context.Context
	cancel       context.CancelFunc
	serviceSpecs *sync.Map
	dbConn       *basenine.Connection
	dbMutex      sync.Mutex
	entriesQuery string
}

func GetDefaultOasGeneratorInstance() *defaultOasGenerator {
	syncOnce.Do(func() {
		instance = NewDefaultOasGenerator()
		logger.Log.Debug("OAS Generator Initialized")
	})
	return instance
}

func (g *defaultOasGenerator) Start(conn *basenine.Connection) {
	if g.started {
		return
	}

	if g.dbConn == nil {
		if conn == nil {
			logger.Log.Infof("Creating new DB connection for OAS generator to address %s:%s", shared.BasenineHost, shared.BaseninePort)
			newConn, err := basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
			if err != nil {
				logger.Log.Error("Error connecting to DB for OAS generator, err: %v", err)
				return
			}

			conn = newConn
		}

		g.dbConn = conn
	}

	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel
	g.ctx = ctx
	g.serviceSpecs = &sync.Map{}

	g.started = true

	go g.runGenerator()
}

func (g *defaultOasGenerator) Stop() {
	if !g.started {
		return
	}
	g.started = false

	g.cancel()
	g.reset()

	g.dbMutex.Lock()
	defer g.dbMutex.Unlock()
	if g.dbConn != nil {
		g.dbConn.Close()
		g.dbConn = nil
	}
}

func (g *defaultOasGenerator) IsStarted() bool {
	return g.started
}

func (g *defaultOasGenerator) runGenerator() {
	// Make []byte channels to receive the data and the meta
	dataChan := make(chan []byte)
	metaChan := make(chan []byte)

	g.dbMutex.Lock()
	defer g.dbMutex.Unlock()
	logger.Log.Infof("Querying DB for OAS generator with query '%s'", g.entriesQuery)
	if err := g.dbConn.Query(g.entriesQuery, dataChan, metaChan); err != nil {
		logger.Log.Errorf("Query mode call failed: %v", err)
	}

	for {
		select {
		case <-g.ctx.Done():
			logger.Log.Infof("OAS Generator was canceled")
			close(dataChan)
			close(metaChan)
			return

		case metaBytes, ok := <-metaChan:
			if !ok {
				logger.Log.Infof("OAS Generator - meta channel closed")
				break
			}
			logger.Log.Debugf("Meta: %s", metaBytes)

		case dataBytes, ok := <-dataChan:
			if !ok {
				logger.Log.Infof("OAS Generator - entries channel closed")
				break
			}

			logger.Log.Debugf("Data: %s", dataBytes)
			e := new(api.Entry)
			err := json.Unmarshal(dataBytes, e)
			if err != nil {
				continue
			}
			g.handleEntry(e)
		}
	}
}

func (g *defaultOasGenerator) handleEntry(mizuEntry *api.Entry) {
	if mizuEntry.Protocol.Name == "http" {
		dest := mizuEntry.Destination.Name
		if dest == "" {
			logger.Log.Debugf("OAS: Unresolved entry %d", mizuEntry.Id)
			return
		}

		entry, err := har.NewEntry(mizuEntry.Request, mizuEntry.Response, mizuEntry.StartTime, mizuEntry.ElapsedTime)
		if err != nil {
			logger.Log.Warningf("Failed to turn MizuEntry %d into HAR Entry: %s", mizuEntry.Id, err)
			return
		}

		entryWSource := &EntryWithSource{
			Entry:       *entry,
			Source:      mizuEntry.Source.Name,
			Destination: dest,
			Id:          mizuEntry.Id,
		}

		g.handleHARWithSource(entryWSource)
	} else {
		logger.Log.Debugf("OAS: Unsupported protocol in entry %d: %s", mizuEntry.Id, mizuEntry.Protocol.Name)
	}
}

func (g *defaultOasGenerator) handleHARWithSource(entryWSource *EntryWithSource) {
	entry := entryWSource.Entry
	gen := g.getGen(entryWSource.Destination, entry.Request.URL)

	opId, err := gen.feedEntry(entryWSource)
	if err != nil {
		txt, suberr := json.Marshal(entry)
		if suberr == nil {
			logger.Log.Debugf("Problematic entry: %s", txt)
		}

		logger.Log.Warningf("Failed processing entry %d: %s", entryWSource.Id, err)
		return
	}

	logger.Log.Debugf("Handled entry %s as opId: %s", entryWSource.Id, opId) // TODO: set opId back to entry?
}

func (g *defaultOasGenerator) getGen(dest string, urlStr string) *SpecGen {
	u, err := url.Parse(urlStr)
	if err != nil {
		logger.Log.Errorf("Failed to parse entry URL: %v, err: %v", urlStr, err)
	}

	val, found := g.serviceSpecs.Load(dest)
	var gen *SpecGen
	if !found {
		gen = NewGen(u.Scheme + "://" + dest)
		g.serviceSpecs.Store(dest, gen)
	} else {
		gen = val.(*SpecGen)
	}
	return gen
}

func (g *defaultOasGenerator) reset() {
	g.serviceSpecs = &sync.Map{}
}

func (g *defaultOasGenerator) GetServiceSpecs() *sync.Map {
	return g.serviceSpecs
}

func (g *defaultOasGenerator) SetEntriesQuery(query string) bool {
	changed := g.entriesQuery != query
	g.entriesQuery = query
	return changed
}

func NewDefaultOasGenerator() *defaultOasGenerator {
	return &defaultOasGenerator{
		started:      false,
		ctx:          nil,
		cancel:       nil,
		serviceSpecs: nil,
		dbConn:       nil,
	}
}

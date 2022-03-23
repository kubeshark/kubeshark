package oas

import (
	"context"
	"encoding/json"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/tap/api"
	"net/url"
	"sync"

	"github.com/up9inc/mizu/shared/logger"
)

var (
	syncOnce sync.Once
	instance *defaultOasGenerator
)

type OasGeneratorSink interface {
	PushEntry(entryWithSource *EntryWithSource)
}

type OasGenerator interface {
	Start()
	Stop()
	IsStarted() bool
	Reset()
	GetServiceSpecs() *sync.Map
}

type defaultOasGenerator struct {
	started      bool
	ctx          context.Context
	cancel       context.CancelFunc
	serviceSpecs *sync.Map
}

func GetDefaultOasGeneratorInstance() *defaultOasGenerator {
	syncOnce.Do(func() {
		instance = NewDefaultOasGenerator()
		logger.Log.Debug("OAS Generator Initialized")
	})
	return instance
}

func (g *defaultOasGenerator) Start() {
	if g.started {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel
	g.ctx = ctx
	g.serviceSpecs = &sync.Map{}
	g.started = true
	var db *basenine.Connection
	go g.runGenerator(db)
}

func (g *defaultOasGenerator) Stop() {
	if !g.started {
		return
	}
	g.cancel()
	g.Reset()
	g.started = false
}

func (g *defaultOasGenerator) IsStarted() bool {
	return g.started
}

func (g defaultOasGenerator) runGenerator(db *basenine.Connection) {

}

func (g *defaultOasGenerator) runGenerator(connection *basenine.Connection) {
	if connection == nil {
		c, err := basenine.NewConnection(shared.BasenineHost, shared.BaseninePort)
		connection = c
		if err != nil {
			panic(err)
		}
	}

	// Make []byte channels to recieve the data and the meta
	dataChan := make(chan []byte)
	metaChan := make(chan []byte)

	connection.Query("", dataChan, metaChan)

	for {
		select {
		case <-g.ctx.Done():
			logger.Log.Infof("OAS Generator was canceled")
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
			handleEntry(e, g.serviceSpecs)
		}
	}
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

func (g *defaultOasGenerator) Reset() {
	g.serviceSpecs = &sync.Map{}
}

func (g *defaultOasGenerator) GetServiceSpecs() *sync.Map {
	return g.serviceSpecs
}

func NewDefaultOasGenerator() *defaultOasGenerator {
	return &defaultOasGenerator{
		started:      false,
		ctx:          nil,
		cancel:       nil,
		serviceSpecs: nil,
	}
}

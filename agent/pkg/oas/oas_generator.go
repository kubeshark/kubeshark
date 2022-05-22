package oas

import (
	"encoding/json"
	"net/url"
	"sync"

	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/tap/api"

	"github.com/up9inc/mizu/logger"
)

var (
	syncOnce sync.Once
	instance *defaultOasGenerator
)

type OasGeneratorSink interface {
	HandleEntry(mizuEntry *api.Entry)
}

type OasGenerator interface {
	Start()
	Stop()
	IsStarted() bool
	GetServiceSpecs() *sync.Map
}

type defaultOasGenerator struct {
	started      bool
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
	g.started = true
}

func (g *defaultOasGenerator) Stop() {
	if !g.started {
		return
	}

	g.started = false

	g.reset()
}

func (g *defaultOasGenerator) IsStarted() bool {
	return g.started
}

func (g *defaultOasGenerator) HandleEntry(mizuEntry *api.Entry) {
	if !g.started {
		return
	}

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

func NewDefaultOasGenerator() *defaultOasGenerator {
	return &defaultOasGenerator{
		started:      false,
		serviceSpecs: &sync.Map{},
	}
}

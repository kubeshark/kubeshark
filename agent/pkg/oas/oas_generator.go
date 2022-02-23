package oas

import (
	"context"
	"encoding/json"
	"net/url"
	"sync"

	"github.com/up9inc/mizu/agent/pkg/har"
	"github.com/up9inc/mizu/shared/logger"
)

var (
	syncOnce sync.Once
	instance *oasGenerator
)

func GetOasGeneratorInstance() *oasGenerator {
	syncOnce.Do(func() {
		instance = newOasGenerator()
		logger.Log.Debug("OAS Generator Initialized")
	})
	return instance
}

func (g *oasGenerator) Start() {
	if g.started {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel
	g.ctx = ctx
	g.entriesChan = make(chan EntryWithSource, 100) // buffer up to 100 entries for OAS processing
	g.ServiceSpecs = &sync.Map{}
	g.started = true
	go instance.runGenerator()
}

func (g *oasGenerator) Stop() {
	if !g.started {
		return
	}
	g.cancel()
	g.Reset()
	g.started = false
}

func (g *oasGenerator) runGenerator() {
	for {
		select {
		case <-g.ctx.Done():
			logger.Log.Infof("OAS Generator was canceled")
			return

		case entryWithSource, ok := <-g.entriesChan:
			if !ok {
				logger.Log.Infof("OAS Generator - entries channel closed")
				break
			}
			entry := entryWithSource.Entry
			u, err := url.Parse(entry.Request.URL)
			if err != nil {
				logger.Log.Errorf("Failed to parse entry URL: %v, err: %v", entry.Request.URL, err)
			}

			val, found := g.ServiceSpecs.Load(entryWithSource.Destination)
			var gen *SpecGen
			if !found {
				gen = NewGen(u.Scheme + "://" + entryWithSource.Destination)
				g.ServiceSpecs.Store(entryWithSource.Destination, gen)
			} else {
				gen = val.(*SpecGen)
			}

			opId, err := gen.feedEntry(entryWithSource)
			if err != nil {
				txt, suberr := json.Marshal(entry)
				if suberr == nil {
					logger.Log.Debugf("Problematic entry: %s", txt)
				}

				logger.Log.Warningf("Failed processing entry: %s", err)
				continue
			}

			logger.Log.Debugf("Handled entry %s as opId: %s", entry.Request.URL, opId) // TODO: set opId back to entry?
		}
	}
}

func (g *oasGenerator) Reset() {
	g.ServiceSpecs = &sync.Map{}
}

func (g *oasGenerator) PushEntry(entryWithSource *EntryWithSource) {
	if !g.started {
		return
	}
	select {
	case g.entriesChan <- *entryWithSource:
	default:
		logger.Log.Warningf("OAS Generator - entry wasn't sent to channel because the channel has no buffer or there is no receiver")
	}
}

func newOasGenerator() *oasGenerator {
	return &oasGenerator{
		started:      false,
		ctx:          nil,
		cancel:       nil,
		ServiceSpecs: nil,
		entriesChan:  nil,
	}
}

type EntryWithSource struct {
	Source      string
	Destination string
	Entry       har.Entry
	Id          uint
}

type oasGenerator struct {
	started      bool
	ctx          context.Context
	cancel       context.CancelFunc
	ServiceSpecs *sync.Map
	entriesChan  chan EntryWithSource
}

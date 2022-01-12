package oas

import (
	"context"
	"encoding/json"
	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared/logger"
	"net/url"
	"sync"
)

var (
	syncOnce sync.Once
	instance *oasGenerator
)

func GetOasGeneratorInstance(enabled bool) *oasGenerator {
	syncOnce.Do(func() {
		instance = newOasGenerator(enabled)

		if enabled {
			go instance.runGeneretor()
		}

		logger.Log.Debug("Oas Generato Initialized")
	})
	return instance
}

func (g *oasGenerator) runGeneretor() {
	for {
		select {
		case <-g.ctx.Done():
			logger.Log.Infof("OAS Generator was canceled")
			return

		case entry, ok := <-g.entriesChan:
			if !ok {
				logger.Log.Infof("OAS Generator - entries channel closed")
				break
			}
			u, err := url.Parse(entry.Request.URL)
			if err != nil {
				logger.Log.Errorf("Failed to parse entry URL: %v, err: %v", entry.Request.URL, err)
			}

			val, found := g.serviceSpecs.Load(u.Host)
			var gen *SpecGen
			if !found {
				gen = NewGen(u.Scheme + "://" + u.Host)
				g.serviceSpecs.Store(u.Host, gen)
			} else {
				gen = val.(*SpecGen)
			}

			opId, err := gen.feedEntry(entry)
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

func (g *oasGenerator) PushEntry(entry *har.Entry) {
	if !g.enabled {
		return
	}
	select {
	case g.entriesChan <- *entry:
	default:
		logger.Log.Warningf("OAS Generator - entry wasn't sent to channel because the channel has no buffer or there is no receiver")
	}
}

func newOasGenerator(enabled bool) *oasGenerator {
	ctx, cancel := context.WithCancel(context.Background())
	return &oasGenerator{
		enabled:      enabled,
		ctx:          ctx,
		cancel:       cancel,
		serviceSpecs: &sync.Map{},
		entriesChan:  make(chan har.Entry, 100), // buffer up to 100 entries for OAS processing
	}
}

type oasGenerator struct {
	enabled      bool
	ctx          context.Context
	cancel       context.CancelFunc
	serviceSpecs *sync.Map
	entriesChan  chan har.Entry
}

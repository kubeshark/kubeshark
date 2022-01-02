package oas

import (
	"encoding/json"
	"github.com/google/martian/har"
	"github.com/up9inc/mizu/shared/logger"
	"net/url"
	"sync"
)

func EntriesToSpecs(entries chan har.Entry, specs *sync.Map) error {
	for {
		entry, ok := <-entries
		if !ok {
			break
		}

		u, err := url.Parse(entry.Request.URL)
		if err != nil {
			return err
		}

		val, found := specs.Load(u.Host)
		var gen *SpecGen
		if !found {
			gen = NewGen(u.Host)
			specs.Store(u.Host, gen)
		} else {
			gen = val.(*SpecGen)
		}

		err = gen.feedEntry(entry)
		if err != nil {
			txt, suberr := json.Marshal(entry)
			if suberr == nil {
				logger.Log.Debugf("Problematic entry: %s", txt)
			}

			logger.Log.Warningf("Failed processing entry: %s", err)
			continue
		}
	}
	return nil
}

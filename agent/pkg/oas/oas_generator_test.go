package oas

import (
	"encoding/json"
	"github.com/up9inc/mizu/agent/pkg/har"
	"sync"
	"testing"
)

func TestOASGen(t *testing.T) {
	gen := new(defaultOasGenerator)
	gen.dbConn = GetFakeDBConn(`{"startedDateTime": "20000101","request": {"url": "https://host/path", "method": "GET"}, "response": {"status": 200}}`)
	gen.serviceSpecs = &sync.Map{}

	e := new(har.Entry)
	err := json.Unmarshal([]byte(`{"startedDateTime": "20000101","request": {"url": "https://host/path", "method": "GET"}, "response": {"status": 200}}`), e)
	if err != nil {
		panic(err)
	}

	ews := &EntryWithSource{
		Destination: "some",
		Entry:       *e,
	}
	gen.Start()
	gen.handleHARWithSource(ews)
	g, ok := gen.serviceSpecs.Load("some")
	if !ok {
		panic("Failed")
	}
	sg := g.(*SpecGen)
	spec, err := sg.GetSpec()
	if err != nil {
		panic(err)
	}
	specText, _ := json.Marshal(spec)
	t.Log(string(specText))

	if !gen.IsStarted() {
		t.Errorf("Should be started")
	}
	gen.Stop()
}

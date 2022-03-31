package basenine

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	HOST     string = "localhost"
	PORT     string = "9099"
	REDACTED string = "[REDACTED]"
)

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func TestLimit(t *testing.T) {
	err := Limit(HOST, PORT, 1000000)
	assert.Nil(t, err)
}

func TestMacro(t *testing.T) {
	err := Macro(HOST, PORT, "chevy", `brand.name == "Chevrolet"`)
	assert.Nil(t, err)
}

func TestInsertionFilter(t *testing.T) {
	err := InsertionFilter(HOST, PORT, `brand.name == "Chevrolet" and redact("year")`)
	assert.Nil(t, err)
}

func TestInsert(t *testing.T) {
	payload := `{"brand":{"name":"Chevrolet"},"model":"Camaro","year":2021}`

	c, err := NewConnection(HOST, PORT)
	assert.Nil(t, err)

	c.InsertMode()
	for index := 0; index < 15000; index++ {
		c.SendText(payload)
	}
}

func TestSingle(t *testing.T) {
	id := 42
	data, err := Single(HOST, PORT, id, "")
	assert.Nil(t, err)

	expected := fmt.Sprintf(`{"brand":{"name":"Chevrolet"},"id":%d,"model":"Camaro","year":"%s"}`, id, REDACTED)
	assert.JSONEq(t, expected, string(data))
}

func TestValidate(t *testing.T) {
	err := Validate(HOST, PORT, `brand.name == "Chevrolet"`)
	assert.Nil(t, err)

	err = Validate(HOST, PORT, `http and request.method == "GET" and request.path != "/example" and (request.query.a > 42 or request.headers["x"] == "y")`)
	assert.Nil(t, err)

	err = Validate(HOST, PORT, `=.=`)
	assert.EqualError(t, err, `1:1: unexpected token "="`)

	err = Validate(HOST, PORT, `request.path[3.14] == "hello"`)
	assert.EqualError(t, err, `1:14: unexpected token "3.14" (expected (<string> | <char> | <rawstring> | "*") "]")`)
}

func TestQuery(t *testing.T) {
	c, err := NewConnection(HOST, PORT)
	assert.Nil(t, err)

	data := make(chan []byte)
	meta := make(chan []byte)

	handleDataChannel := func(wg *sync.WaitGroup, c *Connection, data chan []byte) {
		defer wg.Done()
		index := 0
		for {
			bytes := <-data

			var d map[string]interface{}
			err := json.Unmarshal(bytes, &d)
			assert.Nil(t, err)

			delete(d, "id")

			ret, err := json.Marshal(d)
			assert.Nil(t, err)
			text := string(ret)

			index++
			assert.JSONEq(t, fmt.Sprintf(`{"brand":{"name":"Chevrolet"},"model":"Camaro","year":"%s"}`, REDACTED), text)

			if index > 14000 {
				c.Close()
				return
			}
		}
	}

	handleMetaChannel := func(c *Connection, meta chan []byte) {
		index := 0
		for {
			bytes := <-meta

			var metadata *Metadata
			err := json.Unmarshal(bytes, &metadata)
			assert.Nil(t, err)

			index++
			if index > 14000 {
				return
			}
		}
	}

	var wg sync.WaitGroup
	go handleDataChannel(&wg, c, data)
	go handleMetaChannel(c, meta)
	wg.Add(1)

	c.Query(`chevy`, data, meta)

	if waitTimeout(&wg, 20*time.Second) {
		t.Fatal("Timed out waiting for wait group")
	}
}

func TestFetch(t *testing.T) {
	data, meta, err := Fetch(HOST, PORT, 100, -1, `chevy`, 20, 20*time.Second)
	assert.Nil(t, err)

	assert.Equal(t, `{"current":20,"total":15000,"numberOfWritten":19,"leftOff":80,"truncatedTimestamp":0}`, string(meta))

	i := 0
	for id := 99; id > 80; id-- {
		expected := fmt.Sprintf(`{"brand":{"name":"Chevrolet"},"id":%d,"model":"Camaro","year":"%s"}`, id, REDACTED)
		assert.JSONEq(t, expected, string(data[i]))
		i++
	}
}

func TestFlush(t *testing.T) {
	err := Flush(HOST, PORT)
	assert.Nil(t, err)
}

func TestReset(t *testing.T) {
	err := Reset(HOST, PORT)
	assert.Nil(t, err)
}

func TestTCPConnectionLeak(t *testing.T) {
	for i := 0; i < 10000; i++ {
		err := Validate(HOST, PORT, `brand.name == "Chevrolet"`)
		assert.Nil(t, err)
		time.Sleep(1 * time.Millisecond)
	}
}

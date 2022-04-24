package controllers

import (
	"bytes"
	basenine "github.com/up9inc/basenine/client/go"
	"net"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/up9inc/mizu/agent/pkg/dependency"
	"github.com/up9inc/mizu/agent/pkg/oas"

	"github.com/gin-gonic/gin"
)

func TestGetOASServers(t *testing.T) {
	recorder, c := getRecorderAndContext()

	GetOASServers(c)
	t.Logf("Written body: %s", recorder.Body.String())
}

func TestGetOASAllSpecs(t *testing.T) {
	recorder, c := getRecorderAndContext()

	GetOASAllSpecs(c)
	t.Logf("Written body: %s", recorder.Body.String())
}

func TestGetOASSpec(t *testing.T) {
	recorder, c := getRecorderAndContext()

	c.Params = []gin.Param{{Key: "id", Value: "some"}}

	GetOASSpec(c)
	t.Logf("Written body: %s", recorder.Body.String())
}

type fakeConn struct {
	sendBuffer    *bytes.Buffer
	receiveBuffer *bytes.Buffer
}

func (f fakeConn) Read(p []byte) (int, error)       { return f.sendBuffer.Read(p) }
func (f fakeConn) Write(p []byte) (int, error)      { return f.receiveBuffer.Write(p) }
func (fakeConn) Close() error                       { return nil }
func (fakeConn) LocalAddr() net.Addr                { return nil }
func (fakeConn) RemoteAddr() net.Addr               { return nil }
func (fakeConn) SetDeadline(t time.Time) error      { return nil }
func (fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (fakeConn) SetWriteDeadline(t time.Time) error { return nil }

func getRecorderAndContext() (*httptest.ResponseRecorder, *gin.Context) {
	dummyConn := new(basenine.Connection)
	dummyConn.Conn = fakeConn{
		sendBuffer:    bytes.NewBufferString("\n"),
		receiveBuffer: bytes.NewBufferString("\n"),
	}
	dependency.RegisterGenerator(dependency.OasGeneratorDependency, func() interface{} {
		return oas.GetDefaultOasGeneratorInstance()
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	oas.GetDefaultOasGeneratorInstance().Start(dummyConn)
	oas.GetDefaultOasGeneratorInstance().GetServiceSpecs().Store("some", oas.NewGen("some"))
	return recorder, c
}

package controllers

import (
	"net/http/httptest"
	"testing"

	"github.com/up9inc/mizu/agent/pkg/oas"

	"github.com/gin-gonic/gin"
)

func TestGetOASServers(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	oas.GetDefaultOasGeneratorInstance().Start()
	oas.GetDefaultOasGeneratorInstance().ServiceSpecs.Store("some", oas.NewGen("some"))

	GetOASServers(c)
	t.Logf("Written body: %s", recorder.Body.String())
}

func TestGetOASAllSpecs(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	oas.GetDefaultOasGeneratorInstance().Start()
	oas.GetDefaultOasGeneratorInstance().ServiceSpecs.Store("some", oas.NewGen("some"))

	GetOASAllSpecs(c)
	t.Logf("Written body: %s", recorder.Body.String())
}

func TestGetOASSpec(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	oas.GetDefaultOasGeneratorInstance().Start()
	oas.GetDefaultOasGeneratorInstance().ServiceSpecs.Store("some", oas.NewGen("some"))

	c.Params = []gin.Param{{Key: "id", Value: "some"}}

	GetOASSpec(c)
	t.Logf("Written body: %s", recorder.Body.String())
}

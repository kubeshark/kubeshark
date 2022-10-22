package controllers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/kubeshark/agent/pkg/dependency"
	"github.com/kubeshark/kubeshark/agent/pkg/oas"
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

func getRecorderAndContext() (*httptest.ResponseRecorder, *gin.Context) {
	dependency.RegisterGenerator(dependency.OasGeneratorDependency, func() interface{} {
		return oas.GetDefaultOasGeneratorInstance(-1)
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	oas.GetDefaultOasGeneratorInstance(-1).Start()
	oas.GetDefaultOasGeneratorInstance(-1).GetServiceSpecs().Store("some", oas.NewGen("some"))
	return recorder, c
}

package controllers

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/oas"
	"net/http/httptest"
	"testing"
)

func TestGetOASServers(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	oas.ServiceSpecs.Store("some", oas.NewGen("some"))

	GetOASServers(c)
	t.Logf("Written body: %s", recorder.Body.String())
	return
}

func TestGetOASSpec(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	oas.ServiceSpecs.Store("some", oas.NewGen("some"))

	c.Params = []gin.Param{{Key: "id", Value: "some"}}

	GetOASSpec(c)
	t.Logf("Written body: %s", recorder.Body.String())
	return
}

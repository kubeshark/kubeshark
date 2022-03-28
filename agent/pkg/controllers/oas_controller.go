package controllers

import (
	"net/http"

	"github.com/chanced/openapi"
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/oas"
	"github.com/up9inc/mizu/shared/logger"
)

func GetOASServers(c *gin.Context) {
	m := make([]string, 0)
	oas.GetOasGeneratorInstance().ServiceSpecs.Range(func(key, value interface{}) bool {
		m = append(m, key.(string))
		return true
	})

	c.JSON(http.StatusOK, m)
}

func GetOASSpec(c *gin.Context) {
	res, ok := oas.GetOasGeneratorInstance().ServiceSpecs.Load(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     true,
			"type":      "error",
			"autoClose": "5000",
			"msg":       "Service not found among specs",
		})
		return // exit
	}

	gen := res.(*oas.SpecGen)
	spec, err := gen.GetSpec()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     true,
			"type":      "error",
			"autoClose": "5000",
			"msg":       err,
		})
		return // exit
	}

	c.JSON(http.StatusOK, spec)
}

func GetOASAllSpecs(c *gin.Context) {
	res := map[string]*openapi.OpenAPI{}
	oas.GetOasGeneratorInstance().ServiceSpecs.Range(func(key, value interface{}) bool {
		svc := key.(string)
		gen := value.(*oas.SpecGen)
		spec, err := gen.GetSpec()
		if err != nil {
			logger.Log.Warningf("Failed to obtain spec for service %s: %s", svc, err)
			return true
		}
		res[svc] = spec
		return true
	})
	c.JSON(http.StatusOK, res)
}

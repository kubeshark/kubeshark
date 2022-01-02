package controllers

import (
	"github.com/gin-gonic/gin"
	"mizuserver/pkg/oas"
	"net/http"
)

func GetOASServers(c *gin.Context) {
	m := make([]string, 0)
	oas.ServiceSpecs.Range(func(key, value interface{}) bool {
		m = append(m, key.(string))
		return true
	})

	c.JSON(http.StatusOK, m)
}

func GetOASSpec(c *gin.Context) {
	res, ok := oas.ServiceSpecs.Load(c.Param("id"))
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

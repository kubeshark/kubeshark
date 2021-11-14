package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/shared"
)

type ValidateResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Query   string `json:"query"`
}

func PostValidate(c *gin.Context) {
	query := c.PostForm("query")
	valid := true
	message := ""

	err := basenine.Validate(shared.BasenineHost, shared.BaseninePort, query)
	if err != nil {
		valid = false
		message = err.Error()
	}

	c.JSON(http.StatusOK, ValidateResponse{
		Valid:   valid,
		Message: message,
		Query:   query,
	})
}

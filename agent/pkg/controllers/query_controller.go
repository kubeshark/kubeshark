package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeshark/kubeshark/shared"
	basenine "github.com/up9inc/basenine/client/go"
)

type ValidateResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
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
	})
}

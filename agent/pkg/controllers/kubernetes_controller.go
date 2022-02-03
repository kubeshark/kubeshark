package controllers

import (
	"context"
	"net/http"

	"github.com/up9inc/mizu/agent/pkg/providers/kubernetes"

	"github.com/gin-gonic/gin"
)

func GetNamespaces(c *gin.Context) {
	kubernetesProvider, err := kubernetes.GetKubernetesProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	namespaces, err := kubernetesProvider.ListAllNamespaces(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	namespaceNames := make([]string, len(namespaces))
	for i, namespace := range namespaces {
		namespaceNames[i] = namespace.Name
	}

	c.JSON(http.StatusOK, namespaceNames)
}

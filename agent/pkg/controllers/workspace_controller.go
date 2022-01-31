package controllers

import (
	"errors"
	"mizuserver/pkg/providers/database"
	"mizuserver/pkg/providers/workspace"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateWorkspace(c *gin.Context) {
	requestCreateWorkspace := &workspace.WorkspaceCreateRequest{}

	if err := c.Bind(requestCreateWorkspace); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if newWorkspace, err := workspace.CreateWorkspace(requestCreateWorkspace.Name, requestCreateWorkspace.Namespaces); err != nil {
		if errors.Is(err, &database.ErrorUniqueConstraintViolation{}) {
			c.JSON(http.StatusConflict, gin.H{"error": "a workspace with this name already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, err)
		}
		return
	} else {
		c.JSON(http.StatusOK, newWorkspace)
	}
}

func ListWorkspace(c *gin.Context) {
	if workspaces, err := workspace.ListWorkspaces(); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	} else {
		c.JSON(http.StatusOK, workspaces)
	}
}

func GetWorkspace(c *gin.Context) {
	if workspace, err := workspace.GetWorkspace(c.Param("workspaceId")); err != nil {
		if errors.Is(err, &database.ErrorNotFound{}) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no workspace with this id was found"})
		} else {
			c.JSON(http.StatusInternalServerError, err)
		}
		return
	} else {
		c.JSON(http.StatusOK, workspace)
	}
}

func UpdateWorkspace(c *gin.Context) {
	requestUpdateWorkspace := &workspace.WorkspaceCreateRequest{}

	if err := c.Bind(requestUpdateWorkspace); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if updatedWorkspace, err := workspace.UpdateWorkspace(c.Param("workspaceId"), requestUpdateWorkspace.Name, requestUpdateWorkspace.Namespaces); err != nil {
		if errors.Is(err, &database.ErrorNotFound{}) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no workspace with this id was found"})
		} else if errors.Is(err, &database.ErrorUniqueConstraintViolation{}) {
			c.JSON(http.StatusConflict, gin.H{"error": "a workspace with this name already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, err)
		}
		return
	} else {
		c.JSON(http.StatusOK, updatedWorkspace)
	}
}

func DeleteWorkspace(c *gin.Context) {
	if err := workspace.DeleteWorkspace(c.Param("workspaceId")); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	} else {
		c.JSON(http.StatusOK, "")
	}
}

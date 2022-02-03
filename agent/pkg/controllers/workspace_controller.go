package controllers

import (
	"errors"
	"net/http"

	"github.com/up9inc/mizu/agent/pkg/providers/database"
	"github.com/up9inc/mizu/agent/pkg/providers/tapConfig"
	"github.com/up9inc/mizu/agent/pkg/providers/userRoles"
	"github.com/up9inc/mizu/agent/pkg/providers/workspace"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared/logger"
)

func CreateWorkspace(c *gin.Context) {
	requestCreateWorkspace := &workspace.WorkspaceCreateRequest{}

	if err := c.Bind(requestCreateWorkspace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		if c.Query("linkUser") != "" {
			if err := userRoles.SetUserWorkspaceRole(c.Query("linkUser"), newWorkspace.Id, userRoles.UserRole); err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
		if err := tapConfig.SyncTappingConfigWithWorkspaceNamespaces(); err != nil {
			logger.Log.Errorf("Error while syncing tapping config: %v", err)
		}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		if err := tapConfig.SyncTappingConfigWithWorkspaceNamespaces(); err != nil {
			logger.Log.Errorf("Error while syncing tapping config: %v", err)
		}
		c.JSON(http.StatusOK, updatedWorkspace)
	}
}

func DeleteWorkspace(c *gin.Context) {
	if err := workspace.DeleteWorkspace(c.Param("workspaceId")); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	} else {
		if err := tapConfig.SyncTappingConfigWithWorkspaceNamespaces(); err != nil {
			logger.Log.Errorf("Error while syncing tapping config: %v", err)
		}
		c.JSON(http.StatusOK, "")
	}
}
